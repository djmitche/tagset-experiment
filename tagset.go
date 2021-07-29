package tagset

// A TagSet represents a set of tags, in an efficient fashion.  TagSets
// implicitly de-duplicate the tags they contain.  They are immutable once
// created (in their public API; threadsafe internal mutability may be used).
type TagSet struct {
	// tags contains a duplicate-free list of the tags in this set
	tags []*Tag
}

// NewTagSet creates a new TagSet containing the given tags.
//
// The caller MUST ensure that the set of tags contains no duplicates
func NewTagSetWithoutDuplicates(tags []*Tag) *TagSet {
	return &TagSet{tags}
}

// Union combines two TagSets into one.
func Union(ts1 *TagSet, ts2 *TagSet) *TagSet {
	tags := make([]*Tag, 0, len(ts1.tags)+len(ts2.tags))
	seen := make(map[uint64]struct{})
	for _, t := range ts1.tags {
		tags = append(tags, t)
		seen[t.hash] = struct{}{}
	}
	for _, t := range ts2.tags {
		if _, f := seen[t.hash]; !f {
			tags = append(tags, t)
		}
	}
	return &TagSet{tags}
}

// DisjointUnion combines two TagSets into one with the assumption that the two
// TagSets are disjoint (share no tags in common).
//
// The caller MUST ensure this is the case.
func DisjointUnion(ts1 *TagSet, ts2 *TagSet) *TagSet {
	var tags []*Tag
	tags = append(tags, ts1.tags...)
	tags = append(tags, ts2.tags...)
	return &TagSet{tags}
}

// Return the hash of this set of tags.  This value does not depend on
// order in which the tags were added or duplication of tags.
func (ts *TagSet) Hash() uint64 {
	var h uint64
	for _, t := range ts.tags {
		h ^= t.hash
	}
	return h
}
