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
			// even if indicated by go, there is no context leak here,
			// the context is defered to be closed after the the proxy listener stops (inside the goroutine).
			// TODO: find a cleaner way to solve this (because the static analysis emits a warning
			// it is possible to actually have a leak without notice)
			deviceConnCtx, deviceConnCancel := context.WithCancel(context.Background())
			go func() {
				defer conn.Close()
				select {
				case <- deviceConnCtx.Done():
					return
				}
			}()
			
			headerLengthBuffer := make([]byte, 2)
			n, err := conn.Read(headerLengthBuffer)
			if err!=nil {
				logrus.Warnf("failed to read length: %v", err)
				deviceConnCancel()
				return
			}
			if n < len(headerLengthBuffer) {
				logrus.Warnf("failed to read length: expected %d byte length", len(headerLengthBuffer))
				deviceConnCancel()
				return
			}

			headerLength := binary.BigEndian.Uint16(headerLengthBuffer)

			headerBuffer := make([]byte, headerLength)
			n, err = conn.Read(headerBuffer)
			if err!=nil {
				logrus.Warnf("failed to read header: %v", err)
				deviceConnCancel()
				return
			}
			if n < int(headerLength) {
				logrus.Warnf("failed to read header: expected %d byte header", headerLength)
				deviceConnCancel()
				return
			}

			header := &DiscoverRequestHeader{}
			err = json.Unmarshal(headerBuffer, header)
			if err!=nil {
				logrus.Warnf("failed to deserialize header: %v", err)
				deviceConnCancel()
				return
			}

			controller.Lock()
			defer controller.Unlock()

			port, err := controller.GetPort()
			if err!=nil {
				logrus.Warnf("failed to add device: %v", err)
				deviceConnCancel()
				return
			}

			go func() {
				deviceConnCtx, deviceConnCancel := context.WithCancel(context.Background())
				defer deviceConnCancel()
				
				device := proxy.NewDevice(conn, deviceConnCancel, header.Name, conn.RemoteAddr().String())
				controller.Devices[port] = *device
				
				logrus.Infof("initializing proxy listener for %s %s", device.Name, device.IP)
				err = proxy.StartProxyListener(deviceConnCtx, deviceConnCancel, *device, port)
				if err!=nil {
					logrus.Warnf("%v\n", err)
				}

				controller.Lock()
				defer controller.Unlock()
				delete(controller.Devices, port)
			}()
		}()
	}
}
