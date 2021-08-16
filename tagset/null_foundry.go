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

	ident.Sort(nondup)

	return &TagSet{
		size:          len(nondup),
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

	ident.Sort(tags)

	return &TagSet{
		size:          len(tags),
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

	// ensure t2 is smaller than t1
	t1size, t2size := ts1.size, ts2.size
	if t1size < t2size {
		ts1, ts2 = ts2, ts1
		t1size, t2size = t2size, t1size
	}

	hashH := ts1.hashH
	hashL := ts1.hashL
	clone := make([]ident.Ident, 0, t1size+t2size)
	clone = append(clone, ts1.tags...)
	serialization := make([]byte, 0, len(ts1.serialization)+len(ts2.serialization)+1)
	serialization = append(serialization, ts1.serialization...)

	// take advantage of the fact that both parent tagsets are sorted to
	// only scan them once, eliminating duplicates.
	tags1, tags2 := ts1.tags, ts2.tags
	var i1, i2 int
	keep := func(id ident.Ident) {
		hashH ^= id.HashH()
		hashL ^= id.HashL()
		clone = append(clone, id)
		serialization = append(append(serialization, byte(',')), id.Bytes()...)
	}

	for i1 < t1size && i2 < t2size {
		t1 := tags1[i1]
		t2 := tags2[i2]
		if t1.Equals(t2) {
			// duplicate -> skip
			i2++
			i1++
		} else if t1.Less(t2) {
			i1++
		} else {
			keep(t2)
			i2++
		}
	}

	// anything left in t2 is not duplicated
	for i2 < t2size {
		keep(tags2[i2])
		i2++
	}

	// since we prepended a `,` to every item, remove it if necessary
	if len(serialization) > 0 && serialization[0] == byte(',') {
		serialization = serialization[1:]
	}

	ident.Sort(clone)

	return &TagSet{
		size:          len(clone),
		tags:          clone,
		hashH:         hashH,
		hashL:         hashL,
		serialization: serialization,
	}
}

func (f *NullFoundry) DisjointUnion(ts1 *TagSet, ts2 *TagSet) *TagSet {
	clone := make([]ident.Ident, 0, ts1.size+ts2.size)
	clone = append(clone, ts1.tags...)
	clone = append(clone, ts2.tags...)

	serialization := make([]byte, 0, len(ts1.serialization)+len(ts2.serialization)+1)
	serialization = append(serialization, ts1.serialization...)
	if len(serialization) > 0 {
		serialization = append(serialization, ',')
	}
	serialization = append(serialization, ts2.serialization...)

	ident.Sort(clone)

	return &TagSet{
		size:          len(clone),
		tags:          clone,
		hashH:         ts1.hashH ^ ts2.hashH,
		hashL:         ts1.hashL ^ ts2.hashL,
		serialization: serialization,
	}
}
