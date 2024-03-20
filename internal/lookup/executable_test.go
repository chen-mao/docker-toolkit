package lookup

import (
	"fmt"
	"testing"

	testlog "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestExecutableLocator(t *testing.T) {
	logger, _ := testlog.NewNullLogger()

	testCases := []struct {
		root             string
		paths            []string
		expectedPrefixes []string
	}{
		{
			root:             "",
			expectedPrefixes: []string{""},
		},
		{
			root:             "",
			paths:            []string{"/"},
			expectedPrefixes: []string{"/"},
		},
		{
			root:             "",
			paths:            []string{"/", "/bin"},
			expectedPrefixes: []string{"/", "/bin"},
		},
		{
			root:             "/",
			expectedPrefixes: []string{"/"},
		},
		{
			root:             "/",
			paths:            []string{"/"},
			expectedPrefixes: []string{"/"},
		},
		{
			root:             "/",
			paths:            []string{"/", "/bin"},
			expectedPrefixes: []string{"/", "/bin"},
		},
		{
			root:             "/some/path",
			paths:            []string{"/", "/bin"},
			expectedPrefixes: []string{"/some/path", "/some/path/bin"},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			e := newExecutableLocator(logger, tc.root, tc.paths...)

			require.EqualValues(t, tc.expectedPrefixes, e.prefixes)
		})
	}
}
