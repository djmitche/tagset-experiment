package tagset

import (
	"github.com/djmitche/tagset/ident"
)

// A shared cache for the empty TagSet
var emptyTagSet = &TagSet{
	serialization: []byte{},
}

// A TagSet represents a set of tags, in an efficient fashion.  TagSets
// implicitly de-duplicate the tags they contain.  They are immutable once
// created (in their public API; threadsafe internal mutability may be used).
// A TagSet has a 128-bit hash, represented as two 64-bit halves.  The
// likelihood of hash collisions is considered low enough to ignore.
type TagSet struct {
	// size is the total number of tags in the tagset
	size int

	// tags contains a duplicate-free list of the tags in this set.  This
	// is not necessarily sorted!
	tags []ident.Ident

	// hashH and hashL contain the hash of all tags in the set.  Hashes are
	// computed from tag hashes in a way that is associative and commutative.
	hashH, hashL uint64

	// serialization of this tagset
	serialization []byte
}

// Hash returns the 128-bit hash of this tagset, high word first
func (ts *TagSet) Hash() (uint64, uint64) {
	return ts.hashH, ts.hashL
}

// HashH returns the high 64 bits of this tagset's hash
func (ts *TagSet) HashH() uint64 {
	return ts.hashH
}

// HashL returns the low 64 bits of this tagset's hash
func (ts *TagSet) HashL() uint64 {
	return ts.hashL
}

// Serialization returns the serialization of this tagset. The returned
// value MUST not be modified.
func (ts *TagSet) Serialization() []byte {
	return ts.serialization
}

// has determines whether a tagset contains the given tag
func (ts *TagSet) has(t ident.Ident) bool {
	// TODO: this could be much, much faster!  Maybe build a set on
	// each TagSet as required (with sync.Once)?

	for _, t2 := range ts.tags {
		if t2.Equals(t) {
			return true
		}
	}

	return false
}

// forEach calls the given function once for each tag in the TagSet
func (ts *TagSet) forEach(f func(ident.Ident)) {
	for _, t := range ts.tags {
		f(t)
	}
}
