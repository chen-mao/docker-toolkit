package modifier

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/XDXCT/xdxct-container-toolkit/internal/test"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	testlog "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	root    string
	binPath string
}

var cfg *testConfig

func TestMain(m *testing.M) {
	// TEST SETUP
	// Determine the module root and the test binary path
	var err error
	moduleRoot, err := test.GetModuleRoot()
	if err != nil {
		logrus.Fatalf("error in test setup: could not get module root: %v", err)
	}
	testBinPath := filepath.Join(moduleRoot, "test", "bin")

	// Set the environment variables for the test
	os.Setenv("PATH", test.PrependToPath(testBinPath, moduleRoot))

	// Store the root and binary paths in the test Config
	cfg = &testConfig{
		root:    moduleRoot,
		binPath: testBinPath,
	}

	// RUN TESTS
	exitCode := m.Run()

	os.Exit(exitCode)
}

func TestAddHookModifier(t *testing.T) {
	logger, logHook := testlog.NewNullLogger()

	testHookPath := filepath.Join(cfg.binPath, "xdxct-container-runtime-hook")

	testCases := []struct {
		description   string
		spec          specs.Spec
		expectedError error
		expectedSpec  specs.Spec
	}{
		{
			description: "empty spec adds hook",
			spec:        specs.Spec{},
			expectedSpec: specs.Spec{
				Hooks: &specs.Hooks{
					Prestart: []specs.Hook{
						{
							Path: testHookPath,
							Args: []string{"xdxct-container-runtime-hook", "prestart"},
						},
					},
				},
			},
		},
		{
			description: "spec with empty hooks adds hook",
			spec: specs.Spec{
				Hooks: &specs.Hooks{},
			},
			expectedSpec: specs.Spec{
				Hooks: &specs.Hooks{
					Prestart: []specs.Hook{
						{
							Path: testHookPath,
							Args: []string{"xdxct-container-runtime-hook", "prestart"},
						},
					},
				},
			},
		},
		{
			description: "hook is not replaced",
			spec: specs.Spec{
				Hooks: &specs.Hooks{
					Prestart: []specs.Hook{
						{
							Path: "xdxct-container-runtime-hook",
						},
					},
				},
			},
			expectedSpec: specs.Spec{
				Hooks: &specs.Hooks{
					Prestart: []specs.Hook{
						{
							Path: "xdxct-container-runtime-hook",
						},
					},
				},
			},
		},
		{
			description: "other hooks are not replaced",
			spec: specs.Spec{
				Hooks: &specs.Hooks{
					Prestart: []specs.Hook{
						{
							Path: "some-hook",
						},
					},
				},
			},
			expectedSpec: specs.Spec{
				Hooks: &specs.Hooks{
					Prestart: []specs.Hook{
						{
							Path: "some-hook",
						},
						{
							Path: testHookPath,
							Args: []string{"xdxct-container-runtime-hook", "prestart"},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		logHook.Reset()

		t.Run(tc.description, func(t *testing.T) {

			m := NewStableRuntimeModifier(logger, testHookPath)

			err := m.Modify(&tc.spec)
			if tc.expectedError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.EqualValues(t, tc.expectedSpec, tc.spec)
		})
	}

}
