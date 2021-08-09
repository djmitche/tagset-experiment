package ident

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Per https://www.sderosiaux.com/articles/2017/08/26/the-murmur3-hash-function--hashtables-bloom-filters-hyperloglog/#distribution-and-collisions, these have the same hash
const equalHashA = "c:\\Program Files\\iTunes\\CoreFoundation.resources\\th.lproj\\Error.strings"
const equalHashB = "c:\\Windows\\SysWOW64\\usbperf.dll"

func makeIdent(ident string) Ident {
	bytes := []byte(ident)
	hashH, hashL := hashIdent(bytes)
	return newIdent(bytes, hashH, hashL)
}

func TestEmptyIdent(t *testing.T) {
	ident := makeIdent("")
	require.Equal(t, ident.Bytes(), []byte{})
}

func TestIdentHash(t *testing.T) {
	ident := makeIdent("x:abc")
	expH, expL := hashIdent([]byte("x:abc"))
	require.Equal(t, ident.HashH(), expH)
	require.Equal(t, ident.HashL(), expL)
	gotH, gotL := ident.Hash()
	require.Equal(t, gotH, expH)
	require.Equal(t, gotL, expL)
}

func TestIdentEqual(t *testing.T) {
	ident1 := makeIdent("x:abc")
	ident2 := makeIdent("x:abc")

	require.True(t, ident1.Equals(ident1), "pointer equality")
	require.True(t, ident2.Equals(ident2), "pointer equality")
	require.True(t, ident1.Equals(ident2), "hash equality")
}

func TestIdentBytes(t *testing.T) {
	ident := makeIdent("x:abc")
	require.Equal(t, []byte("x:abc"), ident.Bytes())
}
