package modifier

import (
	"testing"

	"github.com/opencontainers/runtime-spec/specs-go"
	testlog "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestHookRemover(t *testing.T) {
	logger, _ := testlog.NewNullLogger()

	testCases := []struct {
		description   string
		spec          *specs.Spec
		expectedError error
		expectedSpec  *specs.Spec
	}{
		{
			description: "existing hooks are maintained",
			spec: &specs.Spec{
				Hooks: &specs.Hooks{
					Prestart: []specs.Hook{
						{
							Path: "/hook/a",
							Args: []string{"/hook/a", "arga"},
						},
					},
				},
			},
			expectedSpec: &specs.Spec{
				Hooks: &specs.Hooks{
					Prestart: []specs.Hook{
						{
							Path: "/hook/a",
							Args: []string{"/hook/a", "arga"},
						},
					},
				},
			},
		},
		{
			description: "modification removes existing xdxct-container-runtime-hook",
			spec: &specs.Spec{
				Hooks: &specs.Hooks{
					Prestart: []specs.Hook{
						{
							Path: "/path/to/xdxct-container-runtime-hook",
							Args: []string{"/path/to/xdxct-container-runtime-hook", "prestart"},
						},
					},
				},
			},
			expectedSpec: &specs.Spec{
				Hooks: &specs.Hooks{
					Prestart: nil,
				},
			},
		},
		{
			description: "modification removes existing xdxct-container-toolkit",
			spec: &specs.Spec{
				Hooks: &specs.Hooks{
					Prestart: []specs.Hook{
						{
							Path: "/path/to/xdxct-container-toolkit",
							Args: []string{"/path/to/xdxct-container-toolkit", "prestart"},
						},
					},
				},
			},
			expectedSpec: &specs.Spec{
				Hooks: &specs.Hooks{
					Prestart: nil,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			m := xdxctContainerRuntimeHookRemover{logger: logger}

			err := m.Modify(tc.spec)
			if tc.expectedError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.EqualValues(t, tc.expectedSpec, tc.spec)
		})
	}
}
