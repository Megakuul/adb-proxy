package proxy

import (
	"fmt"
	"sync"
)

type DeviceController struct {
	sync.RWMutex
	firstPort uint16
	lastPort uint16
	portLocks map[uint16]bool
	devices map[string]*Device
}

func NewDeviceController(firstPort uint16, lastPort uint16) *DeviceController {
	return &DeviceController{
		firstPort: firstPort,
		lastPort: lastPort,
		portLocks: map[uint16]bool{},
		devices: map[string]*Device{},
	}
}

func (d *DeviceController) GetDevice(addr string) (*Device, bool) {
	d.RLock()
	defer d.RUnlock()
	if dev, exists := d.devices[addr]; exists {
		return dev, true
	} else {
		return nil, false
	}
}

func (d *DeviceController) ListDevices() []*Device {
	d.RLock()
	defer d.RUnlock()
	devices := []*Device{}
	for _, dev := range d.devices {
		devices = append(devices, dev)
	}
	return devices
}

func (d *DeviceController) AddDevice(addr string, device *Device) {
	d.Lock()
	defer d.Unlock()
	d.devices[addr] = device
}

func (d *DeviceController) RemoveDevice(addr string) {
	d.Lock()
	defer d.Unlock()
	delete(d.devices, addr)
}

func (d *DeviceController) ReservePort() (uint16, error) {
	d.Lock()
	defer d.Unlock()
	for i := d.firstPort; i <= d.lastPort; i++ {
		if locked, exists := d.portLocks[i]; !exists || !locked {
			d.portLocks[i] = true
			return i, nil
		}
	}
	return 0, fmt.Errorf("no port available")
}

func (d *DeviceController) ReleasePort(port uint16) {
	d.Lock()
	defer d.Unlock()
	d.portLocks[port] = false
}
