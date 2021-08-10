package tagset

import "github.com/djmitche/tagset/ident"

// A Foundry produces tagsets.  In general, foundries are not threadsafe.
type Foundry interface {
	// NewWithDuplicates creates a new TagSet from a slice of tags that may
	// contain duplicates.
	//
	// The slice is not retained, and the caller may re-use it after passing it
	// to this function.
	NewWithDuplicates(tags []ident.Ident) *TagSet

	// NewWithoutDuplicates creates a new TagSet containing the given tags.
	//
	// The caller MUST ensure that the set of tags contains no duplicates.  The
	// slice of tags may be retained in the tagset and MUST not be modified
	// after passing it to this function.
	NewWithoutDuplicates(tags []ident.Ident) *TagSet

	// Parse generates a TagSet from a buffer containing comma-separated tags.
	// It detects duplicate tags while parsing.  The buffer is not retained,
	// and the caller may re-use it after passing it to this function.
	Parse(foundry ident.Foundry, rawTags []byte) *TagSet

	// Union combines two TagSets into one, handling the case where duplicates
	// exist between the two tagsets.  This is much slower than DisjointUnion,
	// so callers that can otherwise ensure disjointness should prefer
	// DisjointUnion.
	Union(ts1 *TagSet, ts2 *TagSet) *TagSet

	// DisjointUnion combines two TagSets into one with the assumption that the
	// two TagSets are disjoint (share no tags in common).
	//
	// The caller MUST ensure this is the case.
	DisjointUnion(ts1 *TagSet, ts2 *TagSet) *TagSet
}
