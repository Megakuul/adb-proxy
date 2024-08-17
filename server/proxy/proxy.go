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

func StartProxyListener(deviceConnCtx context.Context, deviceConnCancel context.CancelFunc, device Device, timeout time.Duration) error {	
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", device.ProxyPort))
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
		err = listener.(*net.TCPListener).SetDeadline(time.Now().Add(timeout))
		if err!=nil {
			return fmt.Errorf("failed to set timeout: %v", err)
		}
		
		clientConn, err := listener.Accept()
		if err == io.EOF {
			return nil
		} else if err!=nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return fmt.Errorf("timeout for device: %s exceeded", device.Name)
			}
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
			if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
				return false, fmt.Errorf("failed to read incomming request: %v", err)
			}
		}

		device.ConnWriteLock.Lock()
		_, err = device.Conn.Write(buffer[:n])
		if err == io.EOF {
			device.ConnWriteLock.Unlock()
			return true, nil
		} else if err!=nil {
			device.ConnWriteLock.Unlock()
			return true, fmt.Errorf("failed to proxy incomming request: %v", err)
		}
		device.ConnWriteLock.Unlock()
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

		_, err = clientConn.Write(buffer[:n])
		if err == io.EOF {
			return false, nil
		} else if err!=nil {
			return false, fmt.Errorf("failed to proxy incomming response: %v", err)
		}
	}
}
