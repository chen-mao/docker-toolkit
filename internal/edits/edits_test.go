package edits

import (
	"testing"

	"github.com/XDXCT/xdxct-container-toolkit/internal/discover"
	"github.com/stretchr/testify/require"
)

func TestFromDiscovererAllowsMountsToIterate(t *testing.T) {
	edits, err := FromDiscoverer(discover.None{})
	require.NoError(t, err)

	require.Empty(t, edits.Mounts)
}
