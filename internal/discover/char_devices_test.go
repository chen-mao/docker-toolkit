package discover

import (
	"fmt"
	"testing"

	testlog "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	"github.com/XDXCT/xdxct-container-toolkit/internal/lookup"
)

func TestCharDevices(t *testing.T) {
	logger, logHook := testlog.NewNullLogger()

	testCases := []struct {
		description          string
		input                *charDevices
		expectedMounts       []Mount
		expectedMountsError  error
		expectedDevicesError error
		expectedDevices      []Device
	}{
		{
			description: "dev mounts are empty",
			input: (*charDevices)(
				&mounts{
					lookup: &lookup.LocatorMock{
						LocateFunc: func(string) ([]string, error) {
							return []string{"located"}, nil
						},
					},
					required: []string{"required"},
				},
			),
			expectedDevices: []Device{{Path: "located", HostPath: "located"}},
		},
		{
			description:          "dev devices returns error for nil lookup",
			input:                &charDevices{},
			expectedDevicesError: fmt.Errorf("no lookup defined"),
		},
	}

	for _, tc := range testCases {
		logHook.Reset()

		t.Run(tc.description, func(t *testing.T) {
			tc.input.logger = logger

			mounts, err := tc.input.Mounts()
			if tc.expectedMountsError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.ElementsMatch(t, tc.expectedMounts, mounts)

			devices, err := tc.input.Devices()
			if tc.expectedDevicesError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.ElementsMatch(t, tc.expectedDevices, devices)
		})
	}
}
