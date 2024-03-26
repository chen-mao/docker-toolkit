package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDriverCapabilitiesIntersection(t *testing.T) {
	testCases := []struct {
		capabilities          DriverCapabilities
		supportedCapabilities DriverCapabilities
		expectedIntersection  DriverCapabilities
	}{
		{
			capabilities:          none,
			supportedCapabilities: none,
			expectedIntersection:  none,
		},
		{
			capabilities:          all,
			supportedCapabilities: none,
			expectedIntersection:  none,
		},
		{
			capabilities:          all,
			supportedCapabilities: allDriverCapabilities,
			expectedIntersection:  allDriverCapabilities,
		},
		{
			capabilities:          allDriverCapabilities,
			supportedCapabilities: all,
			expectedIntersection:  allDriverCapabilities,
		},
		{
			capabilities:          none,
			supportedCapabilities: all,
			expectedIntersection:  none,
		},
		{
			capabilities:          none,
			supportedCapabilities: DriverCapabilities("cap1"),
			expectedIntersection:  none,
		},
		{
			capabilities:          DriverCapabilities("cap0,cap1"),
			supportedCapabilities: DriverCapabilities("cap1,cap0"),
			expectedIntersection:  DriverCapabilities("cap0,cap1"),
		},
		{
			capabilities:          defaultDriverCapabilities,
			supportedCapabilities: allDriverCapabilities,
			expectedIntersection:  defaultDriverCapabilities,
		},
		{
			capabilities:          DriverCapabilities("compute,compat32,graphics,utility,video,display"),
			supportedCapabilities: DriverCapabilities("compute,compat32,graphics,utility,video,display,ngx"),
			expectedIntersection:  DriverCapabilities("compute,compat32,graphics,utility,video,display"),
		},
		{
			capabilities:          DriverCapabilities("cap1"),
			supportedCapabilities: none,
			expectedIntersection:  none,
		},
		{
			capabilities:          DriverCapabilities("compute,compat32,graphics,utility,video,display,ngx"),
			supportedCapabilities: DriverCapabilities("compute,compat32,graphics,utility,video,display"),
			expectedIntersection:  DriverCapabilities("compute,compat32,graphics,utility,video,display"),
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			intersection := tc.supportedCapabilities.Intersection(tc.capabilities)
			require.EqualValues(t, tc.expectedIntersection, intersection)
		})
	}
}

func TestDriverCapabilitiesList(t *testing.T) {
	testCases := []struct {
		capabilities DriverCapabilities
		expected     []string
	}{
		{
			capabilities: DriverCapabilities(""),
		},
		{
			capabilities: DriverCapabilities("  "),
		},
		{
			capabilities: DriverCapabilities(","),
		},
		{
			capabilities: DriverCapabilities(",cap"),
			expected:     []string{"cap"},
		},
		{
			capabilities: DriverCapabilities("cap,"),
			expected:     []string{"cap"},
		},
		{
			capabilities: DriverCapabilities("cap0,,cap1"),
			expected:     []string{"cap0", "cap1"},
		},
		{
			capabilities: DriverCapabilities("cap1,cap0,cap3"),
			expected:     []string{"cap1", "cap0", "cap3"},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			require.EqualValues(t, tc.expected, tc.capabilities.list())
		})
	}
}
