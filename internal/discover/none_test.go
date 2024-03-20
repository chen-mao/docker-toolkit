package discover

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNone(t *testing.T) {
	d := None{}

	mounts, err := d.Mounts()
	require.NoError(t, err)
	require.Empty(t, mounts)
}
