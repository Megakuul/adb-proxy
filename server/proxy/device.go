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


type Device struct {
	name string
	addr string
	proxyPort uint16
	timeout time.Duration

	ctx context.Context
	cancel context.CancelFunc
	wg *sync.WaitGroup
	
	connReadLock *sync.Mutex
	connWriteLock *sync.Mutex
	conn net.Conn
}

func NewDevice(
	conn net.Conn,
	proxyPort uint16,
	name, addr string,
	timeout time.Duration) *Device {

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer conn.Close()
		select {
		case <- ctx.Done():
			return
		}
	}()
	
	return &Device{
		name: name,
		addr: addr,
		proxyPort: proxyPort,
		timeout: timeout,
		ctx: ctx,
		cancel: cancel,
		wg: &sync.WaitGroup{},
		conn: conn,
		connReadLock: &sync.Mutex{},
		connWriteLock: &sync.Mutex{},
	}
}

func (d* Device) GetName() string {
	return d.name
}

func (d* Device) GetAddr() string {
	return d.addr
}

func (d* Device) GetPort() uint16 {
	return d.proxyPort
}

func (d* Device) Close() {
	d.cancel()
	d.wg.Wait()
}


func (d* Device) proxyClientToDevice(clientConn net.Conn, bufferSize int) (bool, error) {
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

		d.connWriteLock.Lock()
		_, err = d.conn.Write(buffer[:n])
		if err == io.EOF {
			d.connWriteLock.Unlock()
			return true, nil
		} else if err!=nil {
			d.connWriteLock.Unlock()
			return true, fmt.Errorf("failed to proxy incomming request: %v", err)
		}
		d.connWriteLock.Unlock()
	}
}

func (d* Device) proxyDeviceToClient(clientConn net.Conn, bufferSize int) (bool, error) {
	for {
		d.connReadLock.Lock()
		d.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		buffer := make([]byte, bufferSize)
		n, err := d.conn.Read(buffer)
		if err == io.EOF {
			d.connReadLock.Unlock()
			return true, nil
		} else if err!=nil {
			if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
				d.connReadLock.Unlock()
				return true, fmt.Errorf("failed to read incomming response: %v", err)
			}
		}
		d.connReadLock.Unlock()

		_, err = clientConn.Write(buffer[:n])
		if err == io.EOF {
			return false, nil
		} else if err!=nil {
			return false, fmt.Errorf("failed to proxy incomming response: %v", err)
		}
	}
}

func (d* Device) StartProxyListener() error {
	d.wg.Add(1)
	defer d.wg.Done()
	
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", d.proxyPort))
	if err!=nil {
		return fmt.Errorf("failed to initialize proxy: %v", err)
	}

	go func() {
		defer listener.Close()
		select {
		case <- d.ctx.Done():
			return
		}
	}()
	
	for {
		err = listener.(*net.TCPListener).SetDeadline(time.Now().Add(d.timeout))
		if err!=nil {
			return fmt.Errorf("failed to set timeout: %v", err)
		}
		
		clientConn, err := listener.Accept()
		if err == io.EOF {
			return nil
		} else if err!=nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return fmt.Errorf("timeout for device: %s exceeded", d.name)
			}
			return fmt.Errorf("failed to accept connection: %v", err)
		}

		proxyConnCtx, cancelProxyConnCtx := context.WithCancel(d.ctx)
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
			disconnectDevice, err := d.proxyClientToDevice(clientConn, 1024)
			if err!=nil {
				logrus.Warnf("%v\n", err)
			}
			if disconnectDevice {
				d.cancel()
			}
			cancelProxyConnCtx()
			wg.Done()
		}()
		
		go func() {
			disconnectDevice, err := d.proxyDeviceToClient(clientConn, 1024)
			if err!=nil {
				logrus.Warnf("%v\n", err)
			}
			if disconnectDevice {
				d.cancel()
			}
			cancelProxyConnCtx()
			wg.Done()
		}()

		wg.Wait()
	}
}

