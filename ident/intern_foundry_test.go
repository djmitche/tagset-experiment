package ident

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInternFoundry(t *testing.T) {
	f := NewInternFoundry()

	id1 := f.Ident([]byte("aaa"))
	id2 := f.Ident([]byte("aaa"))
	id3 := f.Ident([]byte("bbb"))

	// check raw pointer equality
	require.True(t, &id1[0] == &id2[0])
	require.False(t, &id1[0] == &id3[0])
	require.False(t, &id2[0] == &id3[0])

	require.True(t, id1.Equals(id2))
	require.False(t, id1.Equals(id3))
	require.False(t, id2.Equals(id3))

	require.Equal(t, id1.HashH(), id2.HashH())
	require.NotEqual(t, id1.HashH(), id3.HashH())
	require.NotEqual(t, id2.HashH(), id3.HashH())

	require.Equal(t, id1.HashL(), id2.HashL())
	require.NotEqual(t, id1.HashL(), id3.HashL())
	require.NotEqual(t, id2.HashL(), id3.HashL())

	id4 := f.get(id1.HashH(), id1.HashL())
	require.True(t, &id1[0] == &id4[0])

	id5 := f.get(0x123, 0x456)
	require.Nil(t, id5)
}
