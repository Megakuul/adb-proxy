package discover

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"net"
	"time"

	"github.com/megakuul/adb-proxy/server/proxy"
	"github.com/sirupsen/logrus"
)

type DiscoverRequestHeader struct {
	Name string `json:"name"`
}

func StartDiscoverListener(listener net.Listener, controller *proxy.DeviceController, deviceTimeout time.Duration) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Warnf("failed to accept connection: %v", err)
			continue
		}
		go func() {			
			headerLengthBuffer := make([]byte, 2)
			n, err := conn.Read(headerLengthBuffer)
			if err!=nil {
				logrus.Warnf("failed to read length: %v", err)
				conn.Close()
				return
			}
			if n < len(headerLengthBuffer) {
				logrus.Warnf("failed to read length: expected %d byte length", len(headerLengthBuffer))
				conn.Close()
				return
			}

			headerLength := binary.BigEndian.Uint16(headerLengthBuffer)

			headerBuffer := make([]byte, headerLength)
			n, err = conn.Read(headerBuffer)
			if err!=nil {
				logrus.Warnf("failed to read header: %v", err)
				conn.Close()
				return
			}
			if n < int(headerLength) {
				logrus.Warnf("failed to read header: expected %d byte header", headerLength)
				conn.Close()
				return
			}

			header := &DiscoverRequestHeader{}
			err = json.Unmarshal(headerBuffer, header)
			if err!=nil {
				logrus.Warnf("failed to deserialize header: %v", err)
				conn.Close()
				return
			}

			port, err := controller.ReservePort()
			if err!=nil {
				logrus.Warnf("failed to add device: %v", err)
				conn.Close()
				return
			}

			go func() {
				deviceConnCtx, deviceConnCancel := context.WithCancel(context.Background())
				go func() {
					defer conn.Close()
					select {
					case <- deviceConnCtx.Done():
						return
					}
				}()
				defer deviceConnCancel()

				deviceAddr, _, err := net.SplitHostPort(conn.RemoteAddr().String())
				if err != nil {
					logrus.Warnf("failed to parse device addr: %v\n", err)
					return
				}
				device := proxy.NewDevice(conn, deviceConnCancel, port, header.Name, deviceAddr)

				oldDevice, exists := controller.GetDevice(deviceAddr)
				if exists {
					oldDevice.CancelFunc()
				}
				
				controller.AddDevice(deviceAddr, device)
				
				logrus.Infof("initializing proxy listener for %s %s", device.Name, device.Addr)
				err = proxy.StartProxyListener(deviceConnCtx, deviceConnCancel, *device, deviceTimeout)
				if err!=nil {
					logrus.Warnf("%v\n", err)
				}

				controller.RemoveDevice(deviceAddr)
				controller.ReleasePort(port)
			}()
		}()
	}
}
