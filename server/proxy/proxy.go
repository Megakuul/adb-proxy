package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type DeviceController struct {
	sync.RWMutex
	FirstPort uint16
	LastPort uint16
	Devices map[uint16]Device
}

func NewDeviceController(firstPort uint16, lastPort uint16) *DeviceController {
	return &DeviceController{
		FirstPort: firstPort,
		LastPort: lastPort,
		Devices: map[uint16]Device{},
	}
}

type Device struct {
	Conn net.Conn
	Name string
	IP string
	CancelFunc context.CancelFunc
}

func NewDevice(conn net.Conn, cancelFunc context.CancelFunc, name, ip string) *Device {
	return &Device{
		Conn: conn,
		Name: name,
		IP: ip,
		CancelFunc: cancelFunc,
	}
}

func (d *DeviceController) GetPort() (uint16, error) {
	for i := d.FirstPort; i <= d.LastPort; i++ {
		if _, exists := d.Devices[i]; !exists {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no port available")
}

func StartProxyListener(deviceConnCtx context.Context, deviceConnCancel context.CancelFunc, device Device, port uint16) error {	
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err!=nil {
		device.Conn.Close()
		return fmt.Errorf("failed to initialize proxy: %v", err)
	}

	go func() {
		defer listener.Close()
		select {
		case <- deviceConnCtx.Done():
			return
		}
	}()
	
	for {
		println("waiting to accept")
		clientConn, err := listener.Accept()
		if err!=nil {
			return fmt.Errorf("failed to accept connection: %v", err)
		}

		proxyConnCtx, cancelProxyConnCtx := context.WithCancel(deviceConnCtx)
		defer cancelProxyConnCtx()
		
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer clientConn.Close()
			select {
			case <-proxyConnCtx.Done():
				return
			}
		}()
		
		go func() {
			disconnectDevice, err := proxyClientToDevice(clientConn, device.Conn, 1024)
			if err!=nil {
				logrus.Warnf("%v\n", err)
			}
			if disconnectDevice {
				deviceConnCancel()
			}
			cancelProxyConnCtx()
			println("proxyClientToDevice is done")
			wg.Done()
		}()
		
		go func() {
			disconnectDevice, err := proxyDeviceToClient(device.Conn, clientConn, 1024)
			if err!=nil {
				logrus.Warnf("%v\n", err)
			}
			if disconnectDevice {
				deviceConnCancel()
			}
			cancelProxyConnCtx()
			println("proxyDeviceToClient is done")
			wg.Done()
		}()

		println("im waiting")
		wg.Wait()
		println("im done")
	}
}

func proxyClientToDevice(clientConn net.Conn, deviceConn net.Conn, bufferSize int) (bool, error) {
	for {
		buffer := make([]byte, bufferSize)
		n, err := clientConn.Read(buffer)
		if err == io.EOF {
			println("eof from client")
			return false, nil
		} else if err!=nil {
			return false, fmt.Errorf("failed to read incomming request: %v", err)
		}

		deviceConn.SetWriteDeadline(time.Now().Add(1 * time.Second))
		_, err = deviceConn.Write(buffer[:n])
		if err == io.EOF {
			return true, nil
		} else if err!=nil {
			if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
				return true, fmt.Errorf("failed to proxy incomming request: %v", err)
			}
		}
	}
}

func proxyDeviceToClient(deviceConn net.Conn, clientConn net.Conn, bufferSize int) (bool, error) {
	for {
		deviceConn.SetReadDeadline(time.Now().Add(1 * time.Second))
		buffer := make([]byte, bufferSize)
		n, err := deviceConn.Read(buffer)
		if err == io.EOF {
			return true, nil
		} else if err!=nil {
			if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
				return true, fmt.Errorf("failed to read incomming response: %v", err)
			}
		} 

		_, err = clientConn.Write(buffer[:n])
		if err == io.EOF {
			println("eof from client")
			return false, nil
		} else if err!=nil {
			return false, fmt.Errorf("failed to proxy incomming response: %v", err)
		}
	}
}
