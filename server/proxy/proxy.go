package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

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
	Port string
	Name string
	IP string
	CancelFunc context.CancelFunc
}

func NewDevice(cancelFunc context.CancelFunc, name, ip, port string) *Device {
	return &Device{
		Port: port,
		Name: name,
		IP: ip,
		CancelFunc: cancelFunc,
	}
}

func (d *DeviceController) GetPort() (uint16, error) {
	for i := d.FirstPort; i <= d.LastPort; i++ {
		if _, exists := d.Devices[i]; exists {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no port available")
}

func StartProxyListener(ctx context.Context, device Device, port uint16) error {
	proxyListenerCtx, cancelProxyListenerCtx := context.WithCancel(ctx)
	defer cancelProxyListenerCtx()
	
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err!=nil {
		return fmt.Errorf("failed to initialize proxy: %v", err)
	}

	go func() {
		defer listener.Close()
		select {
		case <- proxyListenerCtx.Done():
			return
		}
	}()
	
	for {
		clientConn, err := listener.Accept()
		if err!=nil {
			return fmt.Errorf("failed to accept connection: %v", err)
		}
		
		deviceConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", device.IP, device.Port))
		if err!=nil {
			return fmt.Errorf("failed to connect to device: %v", err)
		}

		proxyConnectionCtx, cancelProxyConnectionCtx := context.WithCancel(proxyListenerCtx)

		go func() {
			var wg sync.WaitGroup
			wg.Add(2)

			defer cancelProxyConnectionCtx()

			go func() {
				defer clientConn.Close()
				defer deviceConn.Close()
				select {
				case <-proxyConnectionCtx.Done():
					return
				}
			}()
			
			go func() {
				if err := proxyClientToDevice(clientConn, deviceConn, 1024); err!=nil {
					logrus.Warnf("%v\n", err)
				}
				cancelProxyConnectionCtx()
				wg.Done()
			}()
			
			go func() {
				if err := proxyDeviceToClient(deviceConn, clientConn, 1024); err!=nil {
					logrus.Warnf("%v\n", err)
				}
				cancelProxyConnectionCtx()
				wg.Done()
			}()

			wg.Wait()
		}()
	}
}

func proxyClientToDevice(clientConn net.Conn, deviceConn net.Conn, bufferSize int) error {
	for {
		buffer := make([]byte, bufferSize)
		n, err := clientConn.Read(buffer)
		if err == io.EOF {
			return nil
		} else if err!=nil {
			return fmt.Errorf("failed to read incomming request: %v", err)
		}

		_, err = deviceConn.Write(buffer[:n])
		if err == io.EOF {
			return nil
		} else if err!=nil {
			return fmt.Errorf("failed to proxy incomming request: %v", err)
		}
	}
}

func proxyDeviceToClient(deviceConn net.Conn, clientConn net.Conn, bufferSize int) error {
	for {
		buffer := make([]byte, bufferSize)
		n, err := deviceConn.Read(buffer)
		if err == io.EOF {
			return nil
		} else if err!=nil {
			return fmt.Errorf("failed to read incomming response: %v", err)
		}

		_, err = clientConn.Write(buffer[:n])
		if err == io.EOF {
			return nil
		} else if err!=nil {
			return fmt.Errorf("failed to proxy incomming response: %v", err)
		}
	}
}
