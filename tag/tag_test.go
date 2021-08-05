package tag

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTagHash(t *testing.T) {
	tg := New("x:abc")
	require.Equal(t, hashTag([]byte("x:abc")), tg.Hash())
	require.Equal(t, []byte("x:abc"), tg.Bytes())
}
