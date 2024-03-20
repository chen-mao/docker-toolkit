package main

import (
	"encoding/json"
	"testing"

	"github.com/XDXCT/xdxct-container-toolkit/pkg/config/engine/docker"
	"github.com/stretchr/testify/require"
)

func TestUpdateConfigDefaultRuntime(t *testing.T) {
	const runtimeDir = "/test/runtime/dir"

	testCases := []struct {
		setAsDefault               bool
		runtimeName                string
		expectedDefaultRuntimeName interface{}
	}{
		{},
		{
			setAsDefault:               false,
			expectedDefaultRuntimeName: nil,
		},
		{
			setAsDefault:               true,
			runtimeName:                "NAME",
			expectedDefaultRuntimeName: "NAME",
		},
		{
			setAsDefault:               true,
			runtimeName:                "xdxct-experimental",
			expectedDefaultRuntimeName: "xdxct-experimental",
		},
		{
			setAsDefault:               true,
			runtimeName:                "xdxct",
			expectedDefaultRuntimeName: "xdxct",
		},
	}

	for i, tc := range testCases {
		o := &options{
			setAsDefault: tc.setAsDefault,
			runtimeName:  tc.runtimeName,
			runtimeDir:   runtimeDir,
		}

		config := docker.Config(map[string]interface{}{})

		err := UpdateConfig(&config, o)
		require.NoError(t, err, "%d: %v", i, tc)

		defaultRuntimeName := config["default-runtime"]
		require.EqualValues(t, tc.expectedDefaultRuntimeName, defaultRuntimeName, "%d: %v", i, tc)
	}
}

func TestUpdateConfig(t *testing.T) {
	const runtimeDir = "/test/runtime/dir"

	testCases := []struct {
		config         docker.Config
		setAsDefault   bool
		runtimeName    string
		expectedConfig map[string]interface{}
	}{
		{
			config:       map[string]interface{}{},
			setAsDefault: false,
			expectedConfig: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"xdxct": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
					"xdxct-experimental": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.experimental",
						"args": []string{},
					},
					"xdxct-cdi": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.cdi",
						"args": []string{},
					},
					"xdxct-legacy": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.legacy",
						"args": []string{},
					},
				},
			},
		},
		{
			config:       map[string]interface{}{},
			setAsDefault: false,
			runtimeName:  "NAME",
			expectedConfig: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"NAME": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
					"xdxct-experimental": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.experimental",
						"args": []string{},
					},
					"xdxct-cdi": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.cdi",
						"args": []string{},
					},
					"xdxct-legacy": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.legacy",
						"args": []string{},
					},
				},
			},
		},
		{
			config:       map[string]interface{}{},
			setAsDefault: false,
			runtimeName:  "xdxct-experimental",
			expectedConfig: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"xdxct": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
					"xdxct-experimental": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.experimental",
						"args": []string{},
					},
					"xdxct-cdi": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.cdi",
						"args": []string{},
					},
					"xdxct-legacy": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.legacy",
						"args": []string{},
					},
				},
			},
		},
		{
			config: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"xdxct": map[string]interface{}{
						"path": "xdxct-container-runtime",
						"args": []string{},
					},
				},
			},
			setAsDefault: false,
			expectedConfig: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"xdxct": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
					"xdxct-experimental": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.experimental",
						"args": []string{},
					},
					"xdxct-cdi": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.cdi",
						"args": []string{},
					},
					"xdxct-legacy": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.legacy",
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
			expectedConfig: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"not-xdxct": map[string]interface{}{
						"path": "some-other-path",
						"args": []string{},
					},
					"xdxct": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
					"xdxct-experimental": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.experimental",
						"args": []string{},
					},
					"xdxct-cdi": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.cdi",
						"args": []string{},
					},
					"xdxct-legacy": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.legacy",
						"args": []string{},
					},
				},
			},
		},
		{
			config: map[string]interface{}{
				"default-runtime": "runc",
			},
			setAsDefault: true,
			runtimeName:  "xdxct",
			expectedConfig: map[string]interface{}{
				"default-runtime": "xdxct",
				"runtimes": map[string]interface{}{
					"xdxct": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
					"xdxct-experimental": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.experimental",
						"args": []string{},
					},
					"xdxct-cdi": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.cdi",
						"args": []string{},
					},
					"xdxct-legacy": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.legacy",
						"args": []string{},
					},
				},
			},
		},
		{
			config: map[string]interface{}{
				"default-runtime": "runc",
			},
			setAsDefault: true,
			runtimeName:  "xdxct-experimental",
			expectedConfig: map[string]interface{}{
				"default-runtime": "xdxct-experimental",
				"runtimes": map[string]interface{}{
					"xdxct": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
					"xdxct-experimental": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.experimental",
						"args": []string{},
					},
					"xdxct-cdi": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.cdi",
						"args": []string{},
					},
					"xdxct-legacy": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.legacy",
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
				"runtimes": map[string]interface{}{
					"xdxct": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
					"xdxct-experimental": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.experimental",
						"args": []string{},
					},
					"xdxct-cdi": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.cdi",
						"args": []string{},
					},
					"xdxct-legacy": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.legacy",
						"args": []string{},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		options := &options{
			setAsDefault: tc.setAsDefault,
			runtimeName:  tc.runtimeName,
			runtimeDir:   runtimeDir,
		}

		err := UpdateConfig(&tc.config, options)
		require.NoError(t, err, "%d: %v", i, tc)

		configContent, err := json.MarshalIndent(tc.config, "", "    ")
		require.NoError(t, err)

		expectedContent, err := json.MarshalIndent(tc.expectedConfig, "", "    ")
		require.NoError(t, err)

		require.EqualValues(t, string(expectedContent), string(configContent), "%d: %v", i, tc)
	}
}

func TestRevertConfig(t *testing.T) {
	testCases := []struct {
		config         docker.Config
		expectedConfig map[string]interface{}
	}{
		{
			config:         map[string]interface{}{},
			expectedConfig: map[string]interface{}{},
		},
		{
			config: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"xdxct": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
				},
			},
			expectedConfig: map[string]interface{}{},
		},
		{
			config: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"xdxct-experimental": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
				},
			},
			expectedConfig: map[string]interface{}{},
		},
		{
			config: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"xdxct": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
					"xdxct-experimental": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.experimental",
						"args": []string{},
					},
				},
			},
			expectedConfig: map[string]interface{}{},
		},
		{
			config: map[string]interface{}{
				"runtimes": map[string]interface{}{
					"xdxct": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
					"xdxct-experimental": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.experimental",
						"args": []string{},
					},
					"xdxct-cdi": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.cdi",
						"args": []string{},
					},
					"xdxct-legacy": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime.legacy",
						"args": []string{},
					},
				},
			},
			expectedConfig: map[string]interface{}{},
		},
		{
			config: map[string]interface{}{
				"default-runtime": "xdxct",
				"runtimes": map[string]interface{}{
					"xdxct": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
				},
			},
			expectedConfig: map[string]interface{}{
				"default-runtime": "runc",
			},
		},
		{
			config: map[string]interface{}{
				"default-runtime": "not-xdxct",
				"runtimes": map[string]interface{}{
					"xdxct": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
				},
			},
			expectedConfig: map[string]interface{}{
				"default-runtime": "not-xdxct",
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
				"runtimes": map[string]interface{}{
					"xdxct": map[string]interface{}{
						"path": "/test/runtime/dir/xdxct-container-runtime",
						"args": []string{},
					},
				},
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
		err := RevertConfig(&tc.config, &options{})

		require.NoError(t, err, "%d: %v", i, tc)

		configContent, err := json.MarshalIndent(tc.config, "", "    ")
		require.NoError(t, err)

		expectedContent, err := json.MarshalIndent(tc.expectedConfig, "", "    ")
		require.NoError(t, err)

		require.EqualValues(t, string(expectedContent), string(configContent), "%d: %v", i, tc)
	}
}
