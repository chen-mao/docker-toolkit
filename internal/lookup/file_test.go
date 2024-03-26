package lookup

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSearchPrefixes(t *testing.T) {
	testCases := []struct {
		root             string
		prefixes         []string
		expectedPrefixes []string
	}{
		{
			root:             "",
			expectedPrefixes: []string{""},
		},
		{
			root:             "/",
			expectedPrefixes: []string{"/"},
		},
		{
			root:             "/some/root",
			expectedPrefixes: []string{"/some/root"},
		},
		{
			root:             "",
			prefixes:         []string{"foo", "bar"},
			expectedPrefixes: []string{"foo", "bar"},
		},
		{
			root:             "/",
			prefixes:         []string{"foo", "bar"},
			expectedPrefixes: []string{"/foo", "/bar"},
		},
		{
			root:             "/",
			prefixes:         []string{"/foo", "/bar"},
			expectedPrefixes: []string{"/foo", "/bar"},
		},
		{
			root:             "/some/root",
			prefixes:         []string{"foo", "bar"},
			expectedPrefixes: []string{"/some/root/foo", "/some/root/bar"},
		},
		{
			root:             "",
			prefixes:         []string{"foo", "bar", "bar", "foo"},
			expectedPrefixes: []string{"foo", "bar"},
		},
		{
			root:             "/some/root",
			prefixes:         []string{"foo", "bar", "foo", "bar"},
			expectedPrefixes: []string{"/some/root/foo", "/some/root/bar"},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			prefixes := getSearchPrefixes(tc.root, tc.prefixes...)
			require.EqualValues(t, tc.expectedPrefixes, prefixes)
		})
	}
}
