package discover

import (
	"testing"

	testlog "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup"
)

func TestIPCMounts(t *testing.T) {
	logger, _ := testlog.NewNullLogger()
	l := ipcMounts(
		mounts{
			logger: logger,
			lookup: &lookup.LocatorMock{
				LocateFunc: func(path string) ([]string, error) {
					return []string{"/host/path"}, nil
				},
			},
			required: []string{"target"},
		},
	)

	mounts, err := l.Mounts()
	require.NoError(t, err)

	require.EqualValues(
		t,
		[]Mount{
			{
				HostPath: "/host/path",
				Path:     "/host/path",
				Options: []string{
					"ro",
					"nosuid",
					"nodev",
					"bind",
					"noexec",
				},
			},
		},
		mounts,
	)
}
