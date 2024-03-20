package image

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseMajorMinorVersionValid(t *testing.T) {
	var tests = []struct {
		version  string
		expected string
	}{
		{"0", "0.0"},
		{"8", "8.0"},
		{"7.5", "7.5"},
		{"9.0.116", "9.0"},
		{"4294967295.4294967295.4294967295", "4294967295.4294967295"},
		{"v11.6", "11.6"},
	}
	for _, c := range tests {
		t.Run(c.version, func(t *testing.T) {
			version, err := parseMajorMinorVersion(c.version)

			require.NoError(t, err)
			require.Equal(t, c.expected, version)
		})
	}
}

func TestParseMajorMinorVersionInvalid(t *testing.T) {
	var tests = []string{
		"foo",
		"foo.5.10",
		"9.0.116.50",
		"9.0.116foo",
		"7.foo",
		"9.0.bar",
		"9.4294967296",
		"9.0.116.",
		"9..0",
		"9.",
		".5.10",
		"-9",
		"+9",
		"-9.1.116",
		"-9.-1.-116",
	}
	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			_, err := parseMajorMinorVersion(c)
			require.Error(t, err)
		})
	}
}

func TestGetRequirements(t *testing.T) {
	testCases := []struct {
		description  string
		env          []string
		requirements []string
	}{
		{
			description:  "XDXCT_REQUIRE_JETPACK is ignored",
			env:          []string{"XDXCT_REQUIRE_JETPACK=csv-mounts=all"},
			requirements: nil,
		},
		{
			description:  "XDXCT_REQUIRE_JETPACK_HOST_MOUNTS is ignored",
			env:          []string{"XDXCT_REQUIRE_JETPACK_HOST_MOUNTS=base-only"},
			requirements: nil,
		},
		{
			description:  "single requirement set",
			env:          []string{"XDXCT_REQUIRE_CUDA=cuda>=11.6"},
			requirements: []string{"cuda>=11.6"},
		},
		{
			description:  "requirements are concatenated requirement set",
			env:          []string{"XDXCT_REQUIRE_CUDA=cuda>=11.6", "XDXCT_REQUIRE_BRAND=brand=tesla"},
			requirements: []string{"cuda>=11.6", "brand=tesla"},
		},
		{
			description:  "legacy image",
			env:          []string{"CUDA_VERSION=11.6"},
			requirements: []string{"cuda>=11.6"},
		},
		{
			description:  "legacy image with additional requirement",
			env:          []string{"CUDA_VERSION=11.6", "XDXCT_REQUIRE_BRAND=brand=tesla"},
			requirements: []string{"cuda>=11.6", "brand=tesla"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			image, err := NewCUDAImageFromEnv(tc.env)
			require.NoError(t, err)

			requirements, err := image.GetRequirements()
			require.NoError(t, err)
			require.ElementsMatch(t, tc.requirements, requirements)

		})

	}
}
