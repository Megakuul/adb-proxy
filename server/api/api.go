package api

import (
	"fmt"
	"net/http"

	"github.com/megakuul/adb-proxy/server/proxy"
)

func StartApiListener(port uint16, controller *proxy.DeviceController) error {
	http.Handle("/", NewListDeviceHandler(controller))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
