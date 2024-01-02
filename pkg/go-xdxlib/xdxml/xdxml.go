package xdxml

import (
	"sync"

	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxml/xdxml"
)

type xdxmlLib struct {
	sync.Mutex
	refcount int
}

var _ Interface = (*xdxmlLib)(nil)

func New() Interface {
	return &xdxmlLib{}
}

func (x *xdxmlLib) Init() Return {
	ret := xdxml.Init()
	if ret != xdxml.SUCCESS {
		return Return(ret)
	}

	x.Lock()
	defer x.Unlock()
	x.refcount++
	return Return(xdxml.SUCCESS)
}

func (x *xdxmlLib) Shutdown() Return {
	ret := xdxml.Shutdown()
	if ret != xdxml.SUCCESS {
		return Return(ret)
	}

	x.Lock()
	defer x.Unlock()
	x.refcount--
	return Return(xdxml.SUCCESS)
}

func (x *xdxmlLib) DeviceGetCount() (int, Return) {
	c, ret := xdxml.DeviceGetCount()
	return c, Return(ret)
}

func (n *xdxmlLib) DeviceGetHandleByIndex(index int) (Device, Return) {
	d, r := xdxml.DeviceGetHandleByIndex(index)
	return xdxmlDevice(d), Return(r)
}
