package docker

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateConfigDefaultRuntime(t *testing.T) {
	testCases := []struct {
		config                     Config
		runtimeName                string
		setAsDefault               bool
		expectedDefaultRuntimeName interface{}
	}{
		{
			setAsDefault:               false,
			expectedDefaultRuntimeName: nil,
		},
		{
			runtimeName:                "NAME",
			setAsDefault:               true,
			expectedDefaultRuntimeName: "NAME",
		},
		{
			config: map[string]interface{}{
				"default-runtime": "ALREADY_SET",
			},
			runtimeName:                "NAME",
			setAsDefault:               false,
			expectedDefaultRuntimeName: "ALREADY_SET",
		},
		{
			config: map[string]interface{}{
				"default-runtime": "ALREADY_SET",
			},
			runtimeName:                "NAME",
			setAsDefault:               true,
			expectedDefaultRuntimeName: "NAME",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			if tc.config == nil {
				tc.config = make(map[string]interface{})
			}
			err := tc.config.AddRuntime(tc.runtimeName, "", tc.setAsDefault)
			require.NoError(t, err)

			defaultRuntimeName := tc.config["default-runtime"]
			require.EqualValues(t, tc.expectedDefaultRuntimeName, defaultRuntimeName)
		})
	}
}

func TestUpdateConfigRuntimes(t *testing.T) {
	testCases := []struct {
		config         Config
		runtimes       map[string]string
		expectedConfig map[string]interface{}
	}{
		{
			config: map[string]interface{}{},
			runtimes: map[string]string{
				"runtime1": "/test/runtime/dir/runtime1",
				"runtime2": "/test/runtime/dir/runtime2",
			},
			expectedConfig: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"runtime1": map[string]interface{}{
						"path": "/test/runtime/dir/runtime1",
						"args": []string{},
					},
					"runtime2": map[string]interface{}{
						"path": "/test/runtime/dir/runtime2",
						"args": []string{},
					},
				},
			},
		},
		{
			config: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"runtime1": map[string]interface{}{
						"path": "runtime1",
						"args": []string{},
					},
				},
			},
			runtimes: map[string]string{
				"runtime1": "/test/runtime/dir/runtime1",
				"runtime2": "/test/runtime/dir/runtime2",
			},
			expectedConfig: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"runtime1": map[string]interface{}{
						"path": "/test/runtime/dir/runtime1",
						"args": []string{},
					},
					"runtime2": map[string]interface{}{
						"path": "/test/runtime/dir/runtime2",
						"args": []string{},
					},
				},
			},
		},
		{
			config: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"not-xdxct": map[string]interface{}{
						"path": "some-other-path",
						"args": []string{},
					},
				},
			},
			runtimes: map[string]string{
				"runtime1": "/test/runtime/dir/runtime1",
			},
			expectedConfig: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"not-xdxct": map[string]interface{}{
						"path": "some-other-path",
						"args": []string{},
					},
					"runtime1": map[string]interface{}{
						"path": "/test/runtime/dir/runtime1",
						"args": []string{},
					},
				},
			},
		},
		{
			config: map[string]interface{}{
				"exec-opts":  []string{"native.cgroupdriver=systemd"},
				"log-driver": "json-file",
				"log-opts": map[string]string{
					"max-size": "100m",
				},
				"storage-driver": "overlay2",
			},
			runtimes: map[string]string{
				"runtime1": "/test/runtime/dir/runtime1",
			},
			expectedConfig: map[string]interface{}{
				"exec-opts":  []string{"native.cgroupdriver=systemd"},
				"log-driver": "json-file",
				"log-opts": map[string]string{
					"max-size": "100m",
				},
				"storage-driver": "overlay2",
				"runtimes": map[string]interface{}{
					"runtime1": map[string]interface{}{
						"path": "/test/runtime/dir/runtime1",
						"args": []string{},
					},
				},
			},
		},
		{
			config: map[string]interface{}{
				"exec-opts":  []string{"native.cgroupdriver=systemd"},
				"log-driver": "json-file",
				"log-opts": map[string]string{
					"max-size": "100m",
				},
				"storage-driver": "overlay2",
			},
			expectedConfig: map[string]interface{}{
				"exec-opts":  []string{"native.cgroupdriver=systemd"},
				"log-driver": "json-file",
				"log-opts": map[string]string{
					"max-size": "100m",
				},
				"storage-driver": "overlay2",
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			for runtimeName, runtimePath := range tc.runtimes {
				err := tc.config.AddRuntime(runtimeName, runtimePath, false)
				require.NoError(t, err)
			}

			configContent, err := json.MarshalIndent(tc.config, "", "    ")
			require.NoError(t, err)

			expectedContent, err := json.MarshalIndent(tc.expectedConfig, "", "    ")
			require.NoError(t, err)

			require.EqualValues(t, string(expectedContent), string(configContent))
		})

	}
}
