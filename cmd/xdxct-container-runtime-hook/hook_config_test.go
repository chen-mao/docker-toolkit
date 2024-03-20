package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetHookConfig(t *testing.T) {
	testCases := []struct {
		lines                      []string
		expectedPanic              bool
		expectedDriverCapabilities DriverCapabilities
	}{
		{
			expectedDriverCapabilities: allDriverCapabilities,
		},
		{
			lines: []string{
				"supported-driver-capabilities = \"all\"",
			},
			expectedDriverCapabilities: allDriverCapabilities,
		},
		{
			lines: []string{
				"supported-driver-capabilities = \"compute,utility,not-compute\"",
			},
			expectedPanic: true,
		},
		{
			lines:                      []string{},
			expectedDriverCapabilities: allDriverCapabilities,
		},
		{
			lines: []string{
				"supported-driver-capabilities = \"\"",
			},
			expectedDriverCapabilities: none,
		},
		{
			lines: []string{
				"supported-driver-capabilities = \"utility,compute\"",
			},
			expectedDriverCapabilities: DriverCapabilities("utility,compute"),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			var filename string
			defer func() {
				if len(filename) > 0 {
					os.Remove(filename)
				}
				configflag = nil
			}()

			if tc.lines != nil {
				configFile, err := os.CreateTemp("", "*.toml")
				require.NoError(t, err)
				defer configFile.Close()

				filename = configFile.Name()
				configflag = &filename

				for _, line := range tc.lines {
					_, err := configFile.WriteString(fmt.Sprintf("%s\n", line))
					require.NoError(t, err)
				}
			}

			var config HookConfig
			getHookConfig := func() {
				c, _ := getHookConfig()
				config = *c
			}

			if tc.expectedPanic {
				require.Panics(t, getHookConfig)
				return
			}

			getHookConfig()

			require.EqualValues(t, tc.expectedDriverCapabilities, config.SupportedDriverCapabilities)
		})
	}
}

func TestGetSwarmResourceEnvvars(t *testing.T) {
	testCases := []struct {
		value    string
		expected []string
	}{
		{
			value:    "nil",
			expected: nil,
		},
		{
			value:    "",
			expected: nil,
		},
		{
			value:    " ",
			expected: nil,
		},
		{
			value:    "single",
			expected: []string{"single"},
		},
		{
			value:    "single ",
			expected: []string{"single"},
		},
		{
			value:    "one,two",
			expected: []string{"one", "two"},
		},
		{
			value:    "one ,two",
			expected: []string{"one", "two"},
		},
		{
			value:    "one, two",
			expected: []string{"one", "two"},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			c := &HookConfig{
				SwarmResource: func() *string {
					if tc.value == "nil" {
						return nil
					}
					return &tc.value
				}(),
			}

			envvars := c.getSwarmResourceEnvvars()
			require.EqualValues(t, tc.expected, envvars)
		})
	}
}
