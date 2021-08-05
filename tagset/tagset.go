package tagset

import (
	"github.com/djmitche/tagset/tag"
)

// A TagSet represents a set of tags, in an efficient fashion.  TagSets
// implicitly de-duplicate the tags they contain.  They are immutable once
// created (in their public API; threadsafe internal mutability may be used).
type TagSet struct {
	// tags contains a duplicate-free list of the tags in this set.  This
	// is not necessarily sorted!
	tags []tag.Tag

	// parents are tagsets that are disjoint from each other and from this
	// struct's tags
	parents [2]*TagSet

	// hash is the hash of all tags in the set.  Hashes are computed from tag
	// hashes in a way that is associative and commutative.
	hash uint64
}

// NewWithoutDuplicates creates a new TagSet containing the given tags.
//
// The caller MUST ensure that the set of tags contains no duplicates.  The
// caller MUST NOT modify the slice of tags after passing it to this function.
func NewWithoutDuplicates(tags []tag.Tag) *TagSet {
	var hash uint64
	for _, t := range tags {
		hash ^= t.Hash()
	}

	return &TagSet{
		tags: tags,
		hash: hash,
	}
}

// New creates a new TagSet from a slice of tags that may contain duplicates.
//
// The slice is not retained, and the caller may re-use it after passing it to
// this function.
func New(tags []tag.Tag) *TagSet {
	seen := map[uint64]struct{}{}
	var hash uint64
	nondup := make([]tag.Tag, 0, len(tags))
	for _, t := range tags {
		h := t.Hash()
		_, found := seen[h]
		if !found {
			nondup = append(nondup, t)
			hash ^= h
		}
	}

	return &TagSet{
		tags: nondup,
		hash: hash,
	}
}

// Union combines two TagSets into one.
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
	hash := ts1.hash

	// insert non-duplicate tags from ts2, updating the hash
	ts2.forEach(func(t tag.Tag) {
		if !ts1.has(t) {
			clone = append(clone, t)
			hash ^= t.Hash()
		}
	})

	return &TagSet{
		tags:    clone,
		parents: [2]*TagSet{ts1, nil},
		hash:    hash,
	}
}

// DisjointUnion combines two TagSets into one with the assumption that the two
// TagSets are disjoint (share no tags in common).
//
// The caller MUST ensure this is the case.
func DisjointUnion(ts1 *TagSet, ts2 *TagSet) *TagSet {
	return &TagSet{
		parents: [2]*TagSet{ts1, ts2},
		hash:    ts1.hash ^ ts2.hash,
	}
}

func (ts *TagSet) Hash() uint64 {
	return ts.hash
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
