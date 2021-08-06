package tagset

import (
	"bytes"

	"github.com/djmitche/tagset/tag"
)

// (used in parsing)
var commaSeparator = []byte(",")

// A guess at tag size (16), to eliminate a few unnecessary reallocations of serializations
const avgTagSize = 16

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
	// tags contains a duplicate-free list of the tags in this set.  This
	// is not necessarily sorted!
	tags []tag.Tag

	// parents are tagsets that are disjoint from each other and from this
	// struct's tags
	parents [2]*TagSet

	// hashH and hashL contain the hash of all tags in the set.  Hashes are
	// computed from tag hashes in a way that is associative and commutative.
	hashH, hashL uint64

	// serialization of this tagset
	serialization []byte
}

// NewWithoutDuplicates creates a new TagSet containing the given tags.
//
// The caller MUST ensure that the set of tags contains no duplicates.  The
// slice of tags is retained in the tagset and MUST not be modified after
// passing it to this function.
func NewWithoutDuplicates(tags []tag.Tag) *TagSet {
	var hashH, hashL uint64
	serialization := make([]byte, 0, len(tags)*avgTagSize)

	for _, t := range tags {
		if len(serialization) > 0 {
			serialization = append(serialization, ',')
		}
		serialization = append(serialization, t.Bytes()...)
		hashH ^= t.HashH()
		hashL ^= t.HashL()
	}

	return &TagSet{
		tags:          tags,
		hashH:         hashH,
		hashL:         hashL,
		serialization: serialization,
	}
}

// New creates a new TagSet from a slice of tags that may contain duplicates.
//
// The slice is not retained, and the caller may re-use it after passing it to
// this function.
func New(tags []tag.Tag) *TagSet {
	seen := map[uint64]struct{}{}
	var hashH, hashL uint64
	serialization := make([]byte, 0, len(tags)*avgTagSize)
	nondup := make([]tag.Tag, 0, len(tags))
	for _, t := range tags {
		hh := t.HashH()
		hl := t.HashL()
		_, found := seen[hh] // TODO: handle hash collision
		if !found {
			nondup = append(nondup, t)
			if len(serialization) > 0 {
				serialization = append(serialization, ',')
			}
			serialization = append(serialization, t.Bytes()...)
			hashH ^= hh
			hashL ^= hl
			seen[hh] = struct{}{}
		}
	}

	return &TagSet{
		tags:          nondup,
		hashH:         hashH,
		hashL:         hashL,
		serialization: serialization,
	}
}

// Parse generates a TagSet from a buffer containing comma-separated tags.  It
// detects duplicate tags propery while parsing.
func Parse(rawTags []byte) *TagSet {
	// TODO: cache TagSets based on the input

	if len(rawTags) == 0 {
		return emptyTagSet
	}

	tagsCount := bytes.Count(rawTags, commaSeparator) + 1
	tags := make([]tag.Tag, tagsCount)

	for i := 0; i < tagsCount-1; i++ {
		tagPos := bytes.Index(rawTags, commaSeparator)
		if tagPos < 0 {
			break
		}
		tags[i] = tag.NewFromBytes(rawTags[:tagPos])
		rawTags = rawTags[tagPos+len(commaSeparator):]
	}
	tags[tagsCount-1] = tag.NewFromBytes(rawTags)

	// TODO: check for duplicates while parsing and use New or NewWithoutDuplicates
	return New(tags)
}

// TODO: shared LRU cache for Union and DisjointUnion by hash, so that union of
// two seen-before tagsets is re-used

// Union combines two TagSets into one, handling the case where duplicates
// exist between the two tagsets.  This is much slower than DisjointUnion, so
// callers that can otherwise ensure disjointness should prefer DisjointUnion.
func Union(ts1 *TagSet, ts2 *TagSet) *TagSet {
	// Because these may not be disjoint, we allocate a new array of tags
	// for the smaller of the two parents, and fill it with non-duplicate

	// ensure t2 is smaller than t1.  We will keep t1 as a parent and
	// deduplicate t2
	t1len, t2len := len(ts1.tags), len(ts2.tags)
	if t1len < t2len {
		ts1, ts2 = ts2, ts1
		t1len, t2len = t2len, t1len
	}

	clone := make([]tag.Tag, 0, t2len)
	hashH := ts1.hashH
	hashL := ts1.hashL
	serialization := make([]byte, 0, len(ts1.serialization)+len(ts2.tags)*avgTagSize)
	serialization = append(serialization, ts1.serialization...)

	// insert non-duplicate tags from ts2, updating the hash
	ts2.forEach(func(t tag.Tag) {
		if !ts1.has(t) {
			clone = append(clone, t)
			if len(serialization) > 0 {
				serialization = append(serialization, ',')
			}
			serialization = append(serialization, t.Bytes()...)
			hashH ^= t.HashH()
			hashL ^= t.HashL()
		}
	})

	return &TagSet{
		tags:          clone,
		parents:       [2]*TagSet{ts1, nil},
		hashH:         hashH,
		hashL:         hashL,
		serialization: serialization,
	}
}

// DisjointUnion combines two TagSets into one with the assumption that the two
// TagSets are disjoint (share no tags in common).
//
// The caller MUST ensure this is the case.
func DisjointUnion(ts1 *TagSet, ts2 *TagSet) *TagSet {
	serialization := make([]byte, 0, len(ts1.serialization)+len(ts2.serialization)+1)
	serialization = append(serialization, ts1.serialization...)
	if len(serialization) > 0 {
		serialization = append(serialization, ',')
	}
	serialization = append(serialization, ts2.serialization...)

	return &TagSet{
		parents:       [2]*TagSet{ts1, ts2},
		hashH:         ts1.hashH ^ ts2.hashH,
		hashL:         ts1.hashL ^ ts2.hashL,
		serialization: serialization,
	}
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
func (ts *TagSet) has(t tag.Tag) bool {
	// TODO: this could be much, much faster!  Maybe build a set on
	// each TagSet as required (with sync.Once)?

	for _, t2 := range ts.tags {
		if t2.Equals(&t) {
			return true
		}
	}

	for _, p := range ts.parents {
		if p != nil && p.has(t) {
			return true
		}
	}

	return false
}

// forEach calls the given function once for each tag in the TagSet
func (ts *TagSet) forEach(f func(tag.Tag)) {
	for _, t := range ts.tags {
		f(t)
	}

	for _, p := range ts.parents {
		if p != nil {
			p.forEach(f)
		}
	}
}
