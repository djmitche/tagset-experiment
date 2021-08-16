package ident

import (
	"encoding/binary"
)

// Implementation Note: identifiers are stored as byte slices, with the first 8
// bytes containing the hash of the tag.

// A tag represents a single tag, with a hash.  Idents are immutable after they
// are created.  Idents have a 128 bit hash, represented as two 64-bit halves
// (HashH and HashL).  The likelihood of hash collisions is considered low
// enough to ignore.
type Ident []byte

// newIdent creates a new tag from a given byte array.  The byte slice
// is no longer referenced after return from this function.  Typically Idents
// will be created via a `Foundry`, and not by this function.
func newIdent(i []byte, hashH, hashL uint64) Ident {
	// TODO: use sync.Pool to store unused slices, and resize as necessary
	clone := make([]byte, hashSize+len(i), hashSize+len(i))
	binary.LittleEndian.PutUint64(clone[:hashSize/2], hashH)
	binary.LittleEndian.PutUint64(clone[hashSize/2:hashSize], hashL)
	copy(clone[hashSize:], i)

	return clone
}

// Hash returns the 128-bit hash of the tag, high word first
func (i Ident) Hash() (uint64, uint64) {
	return binary.LittleEndian.Uint64(i[:hashSize/2]), binary.LittleEndian.Uint64(i[hashSize/2 : hashSize])
}

// HashH returns the high hash of the tag
func (i Ident) HashH() uint64 {
	return binary.LittleEndian.Uint64(i[:hashSize/2])
}

// HashL returns the low hash of the tag
func (i Ident) HashL() uint64 {
	return binary.LittleEndian.Uint64(i[hashSize/2 : hashSize])
}

// Equals performs an approximate equality check, in the sense that it compares
// pointers and, if those are not equal, hashes.  This comparison may have false
// positives (from hash collisions) but not false negatives.
func (i1 Ident) Equals(i2 Ident) bool {
	// if two identifiers begin at the same byte, they are equal; identifiers never use
	// different-lengthed slices of the same underlying array.
	// TODO: use reflect.SliceHeader instead, as this is unnecessarily checking length != 0
	return &i1[0] == &i2[0] || i1.HashH() == i2.HashH() || i1.HashL() == i2.HashL()
}

// Less orders identifiers by their hashes.
func (i1 Ident) Less(i2 Ident) bool {
	if &i1[0] == &i2[0] {
		return false
	}
	return i1.HashH() < i2.HashH() || (i1.HashH() == i2.HashH() && i1.HashL() < i2.HashL())
}

// bytes returns the bytes defining the tag
func (i Ident) Bytes() []byte {
	return i[hashSize:]
}
