package devices

import (
	"golang.org/x/sys/unix"

	"github.com/XDXCT/xdxct-container-toolkit/internal/logger"
)

//go:generate moq -stub -out mknod_mock.go . mknoder
type mknoder interface {
	Mknode(string, int, int) error
}

type mknodLogger struct {
	logger.Interface
}

func (m *mknodLogger) Mknode(path string, major, minor int) error {
	m.Infof("Running: mknod --mode=0666 %s c %d %d", path, major, minor)
	return nil
}

type mknodUnix struct{}

func (m *mknodUnix) Mknode(path string, major, minor int) error {
	err := unix.Mknod(path, unix.S_IFCHR, int(unix.Mkdev(uint32(major), uint32(minor))))
	if err != nil {
		return err
	}
	return unix.Chmod(path, 0666)
}
