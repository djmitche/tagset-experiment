package tagset

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTagHash(t *testing.T) {
	tg := NewTag("x:abc")
	require.Equal(t, hash("x:abc"), tg.hash)
}
