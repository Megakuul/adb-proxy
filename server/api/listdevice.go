package api

import (
	"fmt"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/megakuul/adb-proxy/server/proxy"
)

type ListDeviceHandler struct{
	controller *proxy.DeviceController
}

func NewListDeviceHandler(controller *proxy.DeviceController) *ListDeviceHandler {
	return &ListDeviceHandler{
		controller: controller,
	}
}

type device struct {
	ProxyPort string `json:"proxy_port"`
	DeviceName string `json:"device_name"`
	DeviceAddr string `json:"device_addr"`
}

type listDeviceResponse struct {
	Devices []device `json:"devices"`
}

func (h *ListDeviceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	listDeviceResponse := &listDeviceResponse{}
	for _, dev := range h.controller.ListDevices() {
		listDeviceResponse.Devices = append(listDeviceResponse.Devices, device{
			ProxyPort: strconv.Itoa(int(dev.GetPort())),
			DeviceName: dev.GetName(),
			DeviceAddr: dev.GetAddr(),
		})
	}

	resp, err := json.Marshal(&listDeviceResponse)
	if err!=nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte(
			fmt.Sprintf("failed to parse output: %v", err),
		))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
	return
}
