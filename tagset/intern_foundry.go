package tagset

import (
	"github.com/djmitche/tagset/ident"
	"github.com/twmb/murmur3"
)

// A InternFoundry "interns" tagsets that it has seen before, and returns a
// reference to the existing tagset when it is seen again.  An InternFoundry is
// not threadsafe and must not be accessed concurrently.
type InternFoundry struct {
	// Fallback for when the interned value is not found
	NullFoundry

	// TagSets indexed by the hash of their parse input; notably this is NOT
	// the hash of the TagSet itself.
	byParseHash twoChoice

	Hits, Misses uint64
}

func NewInternFoundry() *InternFoundry {
	return &InternFoundry{
		byParseHash: newTwoChoice(),
	}
}

func (f *InternFoundry) NewWithDuplicates(tags []ident.Ident) *TagSet {
	return f.NullFoundry.NewWithDuplicates(tags)
}

func (f *InternFoundry) NewWithoutDuplicates(tags []ident.Ident) *TagSet {
	return f.NullFoundry.NewWithoutDuplicates(tags)
}

func (f *InternFoundry) Parse(foundry ident.Foundry, rawTags []byte) *TagSet {
	rawHashH, rawHashL := murmur3.Sum128(rawTags)
	existing := f.byParseHash.get(rawHashH, rawHashL)
	if existing != nil {
		f.Hits++
		return existing
	}

	f.Misses++
	fresh := f.NullFoundry.Parse(foundry, rawTags)
	f.byParseHash.insert(rawHashH, rawHashL, fresh)
	return fresh
}

func (f *InternFoundry) Union(ts1 *TagSet, ts2 *TagSet) *TagSet {
	return f.NullFoundry.Union(ts1, ts2)
}

func (f *InternFoundry) DisjointUnion(ts1 *TagSet, ts2 *TagSet) *TagSet {
	return f.NullFoundry.DisjointUnion(ts1, ts2)
}
