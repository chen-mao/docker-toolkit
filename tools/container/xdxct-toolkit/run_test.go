package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseArgs(t *testing.T) {
	testCases := []struct {
		args              []string
		expectedRemaining []string
		expectedRoot      string
		expectedError     error
	}{
		{
			args:              []string{},
			expectedRemaining: []string{},
			expectedRoot:      "",
			expectedError:     nil,
		},
		{
			args:              []string{"app"},
			expectedRemaining: []string{"app"},
		},
		{
			args:              []string{"app", "root"},
			expectedRemaining: []string{"app"},
			expectedRoot:      "root",
		},
		{
			args:              []string{"app", "--flag"},
			expectedRemaining: []string{"app", "--flag"},
		},
		{
			args:              []string{"app", "root", "--flag"},
			expectedRemaining: []string{"app", "--flag"},
			expectedRoot:      "root",
		},
		{
			args:          []string{"app", "root", "not-root", "--flag"},
			expectedError: fmt.Errorf("unexpected positional argument(s) [not-root]"),
		},
		{
			args:          []string{"app", "root", "not-root"},
			expectedError: fmt.Errorf("unexpected positional argument(s) [not-root]"),
		},
		{
			args:          []string{"app", "root", "not-root", "also"},
			expectedError: fmt.Errorf("unexpected positional argument(s) [not-root also]"),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			remaining, root, err := ParseArgs(tc.args)
			if tc.expectedError != nil {
				require.EqualError(t, err, tc.expectedError.Error())
			} else {
				require.NoError(t, err)
			}

			require.ElementsMatch(t, tc.expectedRemaining, remaining)
			require.Equal(t, tc.expectedRoot, root)
		})
	}
}
