package tagset

import (
	"bytes"

	"github.com/djmitche/tagset/ident"
)

// A NullFoundry is the simplest possible Foundry: it just creates TagSets as
// necessary.
type NullFoundry struct{}

func NewNullFoundry() *NullFoundry {
	return &NullFoundry{}
}

func (f *NullFoundry) NewWithDuplicates(tags []ident.Ident) *TagSet {
	seen := seenTracker{}
	var hashH, hashL uint64
	serialization := make([]byte, 0, len(tags)*avgTagSize)
	nondup := make([]ident.Ident, 0, len(tags))
	for _, t := range tags {
		hh := t.HashH()
		hl := t.HashL()
		if !seen.seen(hh, hl) {
			nondup = append(nondup, t)
			if len(serialization) > 0 {
				serialization = append(serialization, ',')
			}
			serialization = append(serialization, t.Bytes()...)
			hashH ^= hh
			hashL ^= hl
		}
	}

	return &TagSet{
		tags:          nondup,
		hashH:         hashH,
		hashL:         hashL,
		serialization: serialization,
	}
}

func (f *NullFoundry) NewWithoutDuplicates(tags []ident.Ident) *TagSet {
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

func (f *NullFoundry) Parse(foundry ident.Foundry, rawTags []byte) *TagSet {
	if len(rawTags) == 0 {
		return emptyTagSet
	}

	tagsCount := bytes.Count(rawTags, commaSeparator) + 1
	tags := make([]ident.Ident, tagsCount)

	for i := 0; i < tagsCount-1; i++ {
		tagPos := bytes.Index(rawTags, commaSeparator)
		if tagPos < 0 {
			break
		}
		tags[i] = foundry.Ident(rawTags[:tagPos])
		rawTags = rawTags[tagPos+len(commaSeparator):]
	}
	tags[tagsCount-1] = foundry.Ident(rawTags)

	// just assume there were duplicates in the parse..
	return f.NewWithDuplicates(tags)
}

func (f *NullFoundry) Union(ts1 *TagSet, ts2 *TagSet) *TagSet {
	// Because these may not be disjoint, we allocate a new array of tags
	// for the smaller of the two parents, and fill it with non-duplicate

	// ensure t2 is smaller than t1.  We will keep t1 as a parent and
	// deduplicate t2
	t1len, t2len := len(ts1.tags), len(ts2.tags)
	if t1len < t2len {
		ts1, ts2 = ts2, ts1
		t1len, t2len = t2len, t1len
	}

	clone := make([]ident.Ident, 0, t2len)
	hashH := ts1.hashH
	hashL := ts1.hashL
	serialization := make([]byte, 0, len(ts1.serialization)+len(ts2.tags)*avgTagSize)
	serialization = append(serialization, ts1.serialization...)

	// insert non-duplicate tags from ts2, updating the hash
	ts2.forEach(func(t ident.Ident) {
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

func (f *NullFoundry) DisjointUnion(ts1 *TagSet, ts2 *TagSet) *TagSet {
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
