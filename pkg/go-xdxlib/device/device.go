package device

import (
	"fmt"

	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxlib/xdxml"
)

type Device xdxml.Device
func (d *devicelib) GetDevices() ([]Device, error) {
	var devs []Device
	err := d.VisitDevices(func(i int, dev Device) error {
		devs = append(devs, dev)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return devs, nil
}

func (d *devicelib) VisitDevices(visit func(int, Device) error) error {
	count, ret := d.xdxml.DeviceGetCount()
	if ret != xdxml.SUCCESS {
		return fmt.Errorf("error getting device count: %v", ret)
	}

	for i := 0; i < count; i++ {
		device, ret := d.xdxml.DeviceGetHandleByIndex(i)
		if ret != xdxml.SUCCESS {
			return fmt.Errorf("error getting device handle for index '%v': %v", i, ret)
		}

		err := visit(i, device)
		if err != nil {
			return fmt.Errorf("error visiting device: %v", err)
		}
	}
	return nil
}
