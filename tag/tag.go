package tag

import (
	"encoding/binary"
)

// Implementation Note: Tags are stored as byte slices, with the first 8 bytes
// containing the hash of the tag.

// A tag represents a single tag, with a hash.  Tags are immutable after they
// are created.
type Tag []byte

// NewFromBytes creates a new tag from a given byte array.  The byte slice
// is no longer referenced after return from this function.
func NewFromBytes(t []byte) Tag {
	hash := hashTag(t)

	// TODO: look up the tag by hash to dedupe

	clone := make([]byte, hashSize+len(t), hashSize+len(t))
	binary.LittleEndian.PutUint64(clone[:hashSize], hash)
	copy(clone[hashSize:], t)
	return clone
}

// New creates a new tag from a given string, computing its hash in the
// process.
func New(t string) Tag {
	// TODO: I think this allocates
	return NewFromBytes([]byte(t))
}

// hash returns the hash of the tag
func (t Tag) Hash() uint64 {
	return binary.LittleEndian.Uint64(t[:hashSize])
}

// Equals performs an approximate equality check, in the sense that it compares
// pointers and, if those are not equal, hashes.  This comparison may have false
// positives (from hash collisions) but not false negatives.
func (t1 *Tag) Equals(t2 *Tag) bool {
	return t1 == t2 || t1.Hash() == t2.Hash()
}

// bytes returns the bytes defining the tag
func (t Tag) Bytes() []byte {
	return t[hashSize:]
}
