package runtime

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	l := NewLogger()

	l.Update("", "debug", nil)

	ll := l.Interface.(*logrus.Logger)
	require.Equal(t, logrus.DebugLevel, ll.Level)

	lp := l.previousLogger.(*logrus.Logger)
	require.Equal(t, logrus.InfoLevel, lp.Level)
}
