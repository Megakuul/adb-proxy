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

func (d *DeviceController) GetPort() (uint16, error) {
	for i := d.FirstPort; i <= d.LastPort; i++ {
		if _, exists := d.Devices[i]; !exists {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no port available")
}

type Device struct {
	ConnReadLock *sync.Mutex
	ConnWriteLock *sync.Mutex
	Conn net.Conn
	Name string
	IP string
	CancelFunc context.CancelFunc
}

func NewDevice(conn net.Conn, cancelFunc context.CancelFunc, name, ip string) *Device {
	return &Device{
		ConnReadLock: &sync.Mutex{},
		ConnWriteLock: &sync.Mutex{},
		Conn: conn,
		Name: name,
		IP: ip,
		CancelFunc: cancelFunc,
	}
}

func (d *Device) KeepaliveDevice(ctx context.Context, interval time.Duration) {
	for {
		select {
		case <- ctx.Done():
			return
		case <- time.After(interval):
			d.ConnReadLock.Lock()
			d.Conn.SetReadDeadline(time.Now().Add(time.Second * 3))
			_, err := d.Conn.Read(make([]byte, 0)) // TODO: 0 byte for whatever reason doesnt return error
			if err == io.EOF {
				d.CancelFunc()
			} else if err!=nil {
				if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
					d.CancelFunc()
				}
			}
			d.ConnReadLock.Unlock()
		}
	}
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
		clientConn, err := listener.Accept()
		if err == io.EOF {
			return nil
		} else if err!=nil {
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
			disconnectDevice, err := proxyClientToDevice(clientConn, device, 1024)
			if err!=nil {
				logrus.Warnf("%v\n", err)
			}
			if disconnectDevice {
				deviceConnCancel()
			}
			cancelProxyConnCtx()
			wg.Done()
		}()
		
		go func() {
			disconnectDevice, err := proxyDeviceToClient(device, clientConn, 1024)
			if err!=nil {
				logrus.Warnf("%v\n", err)
			}
			println("disconnected")
			if disconnectDevice {
				deviceConnCancel()
			}
			cancelProxyConnCtx()
			wg.Done()
		}()

		wg.Wait()
	}
}

func proxyClientToDevice(clientConn net.Conn, device Device, bufferSize int) (bool, error) {
	for {
		clientConn.SetReadDeadline(time.Now().Add(1 * time.Second))
		buffer := make([]byte, bufferSize)
		n, err := clientConn.Read(buffer)
		if err == io.EOF {
			return false, nil
		} else if err!=nil {
			return false, fmt.Errorf("failed to read incomming request: %v", err)
		}

		device.ConnWriteLock.Lock()
		defer device.ConnWriteLock.Unlock()
		device.Conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
		_, err = device.Conn.Write(buffer[:n])
		if err == io.EOF {
			return true, nil
		} else if err!=nil {
			if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
				return true, fmt.Errorf("failed to proxy incomming request: %v", err)
			}
		}
	}
}

func proxyDeviceToClient(device Device, clientConn net.Conn, bufferSize int) (bool, error) {
	for {
		device.ConnReadLock.Lock()
		device.Conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		buffer := make([]byte, bufferSize)
		n, err := device.Conn.Read(buffer)
		if err == io.EOF {
			device.ConnReadLock.Unlock()
			return true, nil
		} else if err!=nil {
			if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
				device.ConnReadLock.Unlock()
				return true, fmt.Errorf("failed to read incomming response: %v", err)
			}
		}
		device.ConnReadLock.Unlock()

		clientConn.SetWriteDeadline(time.Now().Add(1 * time.Second))
		_, err = clientConn.Write(buffer[:n])
		if err == io.EOF {
			return false, nil
		} else if err!=nil {
			return false, fmt.Errorf("failed to proxy incomming response: %v", err)
		}
	}
}
