package discover

import (
	"fmt"
	"testing"

	testlog "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

const (
	testXdxctCTKPath = "/foo/bar/xdxct-ctk"
)

func TestLDCacheUpdateHook(t *testing.T) {
	logger, _ := testlog.NewNullLogger()

	testCases := []struct {
		description   string
		mounts        []Mount
		mountError    error
		expectedError error
		expectedArgs  []string
	}{
		{
			description:  "empty mounts",
			expectedArgs: []string{"xdxct-ctk", "hook", "update-ldcache"},
		},
		{
			description:   "mount error",
			mountError:    fmt.Errorf("mountError"),
			expectedError: fmt.Errorf("mountError"),
		},
		{
			description: "library folders are added to args",
			mounts: []Mount{
				{
					Path: "/usr/local/lib/libfoo.so",
				},
				{
					Path: "/usr/bin/notlib",
				},
				{
					Path: "/usr/local/libother/libfoo.so",
				},
				{
					Path: "/usr/local/lib/libbar.so",
				},
			},
			expectedArgs: []string{"xdxct-ctk", "hook", "update-ldcache", "--folder", "/usr/local/lib", "--folder", "/usr/local/libother"},
		},
		{
			description: "host paths are ignored",
			mounts: []Mount{
				{
					HostPath: "/usr/local/other/libfoo.so",
					Path:     "/usr/local/lib/libfoo.so",
				},
			},
			expectedArgs: []string{"xdxct-ctk", "hook", "update-ldcache", "--folder", "/usr/local/lib"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			mountMock := &DiscoverMock{
				MountsFunc: func() ([]Mount, error) {
					return tc.mounts, tc.mountError
				},
			}
			expectedHook := Hook{
				Path:      testXdxctCTKPath,
				Args:      tc.expectedArgs,
				Lifecycle: "createContainer",
			}

			d, err := NewLDCacheUpdateHook(logger, mountMock, testXdxctCTKPath)
			require.NoError(t, err)

			hooks, err := d.Hooks()
			require.Len(t, mountMock.MountsCalls(), 1)
			require.Len(t, mountMock.DevicesCalls(), 0)
			require.Len(t, mountMock.HooksCalls(), 0)
			if tc.expectedError != nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, hooks, 1)

			require.EqualValues(t, hooks[0], expectedHook)

			devices, err := d.Devices()
			require.NoError(t, err)
			require.Empty(t, devices)

			mounts, err := d.Mounts()
			require.NoError(t, err)
			require.Empty(t, mounts)

		})
	}

}

func TestIsLibName(t *testing.T) {
	testCases := []struct {
		name  string
		isLib bool
	}{
		{
			name:  "",
			isLib: false,
		},
		{
			name:  "lib/not/.so",
			isLib: false,
		},
		{
			name:  "lib.so",
			isLib: false,
		},
		{
			name:  "notlibcuda.so",
			isLib: false,
		},
		{
			name:  "libcuda.so",
			isLib: true,
		},
		{
			name:  "libcuda.so.1",
			isLib: true,
		},
		{
			name:  "libcuda.soNOT",
			isLib: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.isLib, isLibName(tc.name))
		})
	}
}
