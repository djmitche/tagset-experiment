package tag

import (
	"encoding/binary"
	"reflect"
	"unsafe"
)

// Implementation Note: Tags are stored as byte slices, with the first 8 bytes
// containing the hash of the tag.

// A tag represents a single tag, with a hash.  Tags are immutable after they
// are created.  Tags have a 128 bit hash, represented as two 64-bit halves
// (HashH and HashL).  The likelihood of hash collisions is considered low
// enough to ignore.
type Tag []byte

// A cache of existing tags, for re-use
// TODO: some kind of expiration (maybe generations?)
// TODO: threadsafe (sync.Map)
// TODO: handle hash collisions - maybe look up by H then if found but not .Equal, by L?
var tagCache = map[uint64]Tag{}

// NewFromBytes creates a new tag from a given byte array.  The byte slice
// is no longer referenced after return from this function.
func NewFromBytes(t []byte) Tag {
	hashH, hashL := hashTag(t)

	existing, found := tagCache[hashH]
	if found {
		return existing
	}

	// TODO: use sync.Pool to store unused slices, and resize as necessary
	clone := make([]byte, hashSize+len(t), hashSize+len(t))
	binary.LittleEndian.PutUint64(clone[:hashSize/2], hashH)
	binary.LittleEndian.PutUint64(clone[hashSize/2:hashSize], hashL)
	copy(clone[hashSize:], t)

	tagCache[hashH] = clone
	return clone
}

// New creates a new tag from a given string, computing its hash in the
// process.
func New(t string) Tag {
	// This unsafety is to steal the byte slice from the string without allocating.
	// The bytes are copied in NewFromBytes, so this usage is only temporary. See
	// https://stackoverflow.com/questions/68401381/byte-slice-converted-with-unsafe-from-string-changes-its-address
	const max = 1<<31 - 1
	bytes := (*[max]byte)(unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&t)).Data))[:len(t):len(t)]

	return NewFromBytes(bytes)
}

// Hash returns the 128-bit hash of the tag, high word first
func (t Tag) Hash() (uint64, uint64) {
	return binary.LittleEndian.Uint64(t[:hashSize/2]), binary.LittleEndian.Uint64(t[hashSize/2 : hashSize])
}

// HashH returns the high hash of the tag
func (t Tag) HashH() uint64 {
	return binary.LittleEndian.Uint64(t[:hashSize/2])
}

// HashL returns the low hash of the tag
func (t Tag) HashL() uint64 {
	return binary.LittleEndian.Uint64(t[hashSize/2 : hashSize])
}

// Equals performs an approximate equality check, in the sense that it compares
// pointers and, if those are not equal, hashes.  This comparison may have false
// positives (from hash collisions) but not false negatives.
func (t1 *Tag) Equals(t2 *Tag) bool {
	return t1 == t2 || t1.HashH() == t2.HashH() || t1.HashL() == t2.HashL()
}

// bytes returns the bytes defining the tag
func (t Tag) Bytes() []byte {
	return t[hashSize:]
}
