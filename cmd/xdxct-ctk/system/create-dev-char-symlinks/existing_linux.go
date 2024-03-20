package devchar

import "golang.org/x/sys/unix"

func newDeviceNode(d string, stat unix.Stat_t) deviceNode {
	deviceNode := deviceNode{
		path:  d,
		major: unix.Major(stat.Rdev),
		minor: unix.Minor(stat.Rdev),
	}
	return deviceNode
}
