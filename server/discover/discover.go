package discover

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"net"

	"github.com/megakuul/adb-proxy/server/proxy"
	"github.com/sirupsen/logrus"
)

type DiscoverRequestHeader struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

func StartDiscoverListener(listener net.Listener, controller *proxy.DeviceController) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Warnf("failed to accept connection: %v", err)
			continue
		}
		go func() {
			defer conn.Close()
			
			headerLengthBuffer := make([]byte, 2)
			n, err := conn.Read(headerLengthBuffer)
			if err!=nil {
				logrus.Warnf("failed to read length: %v", err)
				return
			}
			if n > len(headerLengthBuffer) {
				logrus.Warnf("failed to read length: expected %d byte length", len(headerLengthBuffer))
				return
			}

			headerLength := binary.BigEndian.Uint16(headerLengthBuffer)

			headerBuffer := make([]byte, headerLength)
			n, err = conn.Read(headerBuffer)
			if err!=nil {
				logrus.Warnf("failed to read header: %v", err)
				return
			}
			if n < int(headerLength) {
				logrus.Warnf("failed to read header: expected %d byte header", headerLength)
				return
			}

			header := &DiscoverRequestHeader{}
			err = json.Unmarshal(headerBuffer, header)
			if err!=nil {
				logrus.Warnf("failed to deserialize header: %v", err)
				return
			}

			controller.Lock()
			defer controller.Unlock()

			port, err := controller.GetPort()
			if err!=nil {
				logrus.Warnf("failed to add device: %v", err)
				return
			}

			ctx, cancel := context.WithCancel(context.Background())

			device := proxy.NewDevice(cancel, header.Name, conn.RemoteAddr().String(), header.Port)
			controller.Devices[port] = *device

			go func() {
				logrus.Infof("initializing proxy listener for %s %s", device.Name, device.IP)
				err = proxy.StartProxyListener(ctx, *device, port)
				if err!=nil {
					logrus.Warnf("%v\n", err)
				}

				controller.Lock()
				defer controller.Unlock()
				controller.Devices[port].CancelFunc()
				delete(controller.Devices, port)
			}()
		}()
	}
}
