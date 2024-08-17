package discover

import (
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
				deviceAddr, _, err := net.SplitHostPort(conn.RemoteAddr().String())
				if err != nil {
					logrus.Warnf("failed to parse device addr: %v\n", err)
					return
				}
				device := proxy.NewDevice(conn, port, header.Name, deviceAddr, deviceTimeout)
				
				oldDevice, exists := controller.GetDevice(deviceAddr)
				if exists {
					oldDevice.Close()
				}
				
				controller.AddDevice(deviceAddr, device)
				
				logrus.Infof("initializing proxy listener (:%d) for %s %s",
					device.GetPort(), device.GetName(), device.GetAddr())
				
				err = device.StartProxyListener()
				if err!=nil {
					logrus.Warnf("%v\n", err)
				}

				device.Close()
				controller.RemoveDevice(deviceAddr)
				controller.ReleasePort(port)
			}()
		}()
	}
}
