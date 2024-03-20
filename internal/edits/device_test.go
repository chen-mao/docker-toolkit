
package edits

import (
	"fmt"
	"testing"

	"github.com/XDXCT/xdxct-container-toolkit/internal/discover"
	"github.com/stretchr/testify/require"
	"tags.cncf.io/container-device-interface/specs-go"
)

func TestDeviceToSpec(t *testing.T) {
	testCases := []struct {
		device   discover.Device
		expected *specs.DeviceNode
	}{
		{
			device: discover.Device{
				Path: "/foo",
			},
			expected: &specs.DeviceNode{
				Path: "/foo",
			},
		},
		{
			device: discover.Device{
				Path:     "/foo",
				HostPath: "/foo",
			},
			expected: &specs.DeviceNode{
				Path: "/foo",
			},
		},
		{
			device: discover.Device{
				Path:     "/foo",
				HostPath: "/not/foo",
			},
			expected: &specs.DeviceNode{
				Path:     "/foo",
				HostPath: "/not/foo",
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			spec, err := device(tc.device).toSpec()
			require.NoError(t, err)
			require.EqualValues(t, tc.expected, spec)
		})
	}
}
