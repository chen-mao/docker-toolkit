package modifier

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAnnotationDevices(t *testing.T) {
	testCases := []struct {
		description     string
		prefixes        []string
		annotations     map[string]string
		expectedDevices []string
		expectedError   error
	}{
		{
			description: "no annotations",
		},
		{
			description: "no matching annotations",
			prefixes:    []string{"not-prefix/"},
			annotations: map[string]string{
				"prefix/foo": "example.com/device=bar",
			},
		},
		{
			description: "single matching annotation",
			prefixes:    []string{"prefix/"},
			annotations: map[string]string{
				"prefix/foo": "example.com/device=bar",
			},
			expectedDevices: []string{"example.com/device=bar"},
		},
		{
			description: "multiple matching annotations",
			prefixes:    []string{"prefix/", "another-prefix/"},
			annotations: map[string]string{
				"prefix/foo":         "example.com/device=bar",
				"another-prefix/bar": "example.com/device=baz",
			},
			expectedDevices: []string{"example.com/device=bar", "example.com/device=baz"},
		},
		{
			description: "multiple matching annotations with duplicate devices",
			prefixes:    []string{"prefix/", "another-prefix/"},
			annotations: map[string]string{
				"prefix/foo":         "example.com/device=bar",
				"another-prefix/bar": "example.com/device=bar",
			},
			expectedDevices: []string{"example.com/device=bar"},
		},
		{
			description: "invalid devices",
			prefixes:    []string{"prefix/"},
			annotations: map[string]string{
				"prefix/foo": "example.com/device",
			},
			expectedError: fmt.Errorf("invalid device %q", "example.com/device"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			devices, err := getAnnotationDevices(tc.prefixes, tc.annotations)
			if tc.expectedError != nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.ElementsMatch(t, tc.expectedDevices, devices)
		})
	}
}
