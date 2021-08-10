package tagset

import (
	"bytes"

	"github.com/djmitche/tagset/ident"
)

// (used in parsing)
var commaSeparator = []byte(",")

// A guess at tag size (16), to eliminate a few unnecessary reallocations of serializations
const avgTagSize = 16

// A NullFoundry is the simplest possible Foundry: it just creates TagSets as
// necessary.

type NullFoundry struct{}

func NewNullFoundry() *NullFoundry {
	return &NullFoundry{}
}

// A seenTracker can be used to track seeing things by their 128-bit hashes
type seenTracker map[uint64][]uint64

// Track the given identifier as seen and return true if it had been seen
// before.
func (seen seenTracker) seen(hashH, hashL uint64) bool {
	// the map uses the high uint64 as a hash table index, with a linear search
	// used to find the low uint64 in the bucket.  Almost every bucket will be
	// one item long.  NOTE: test the bucketing by stripping bits from `hashH`
	// with `hashH &= 0x7`.

	hashBucket, found := seen[hashH]
	if found {
		found = false
		for _, existingL := range hashBucket {
			if existingL == hashL {
				found = true
				break
			}
		}
	} else {
		hashBucket = []uint64{}
	}
	if !found {
		seen[hashH] = append(hashBucket, hashL)
	}

	return found
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
