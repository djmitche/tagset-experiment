package ident

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNullFoundryIdent(t *testing.T) {
	f := NewNullFoundry()

	id1 := f.Ident([]byte("abc:def"))
	id2 := f.Ident([]byte("abc:def"))

	// NullFoundry does not deduplicate, so these are differents slices
	require.True(t, &id1[0] != &id2[0])

	// But otherwise the same
	require.Equal(t, id1.HashL(), id2.HashL())
	require.Equal(t, id1.HashH(), id2.HashH())
	require.True(t, id1.Equals(id2))
}

func TestNullFoundryGet(t *testing.T) {
	f := NewNullFoundry()
	require.Nil(t, f.Get(0x123, 0x456))
}
