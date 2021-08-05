package tagset

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/djmitche/tagset/tag"
	"github.com/stretchr/testify/assert"
	"github.com/twmb/murmur3"
)

func hashOf(tags ...string) uint64 {
	var h uint64
	for _, t := range tags {
		h ^= murmur3.Sum64([]byte(t))
	}
	return h
}

func TestSingleTagHash(t *testing.T) {
	tg := tag.New("x:abc")
	ts := NewWithoutDuplicates([]tag.Tag{tg})
	assert.Equal(t, hashOf("x:abc"), ts.Hash())
}

func TestFromBytes(t *testing.T) {
	tg1 := tag.New("x:abc")
	tg2 := tag.NewFromBytes([]byte("y:def"))
	ts := NewWithoutDuplicates([]tag.Tag{tg1, tg2})
	assert.Equal(t, hashOf("x:abc", "y:def"), ts.Hash())
}

func TestTwoTagHash(t *testing.T) {
	tg1 := tag.New("x:abc")
	tg2 := tag.New("y:def")

	expHash := hashOf("x:abc", "y:def")

	// hash should be the same regardless of order
	ts12 := NewWithoutDuplicates([]tag.Tag{tg1, tg2})
	assert.Equal(t, expHash, ts12.Hash())

	ts21 := NewWithoutDuplicates([]tag.Tag{tg2, tg1})
	assert.Equal(t, expHash, ts21.Hash())
}

func TestDisjointUnions(t *testing.T) {
	test := func(union func(t1 *TagSet, t2 *TagSet) *TagSet) func(*testing.T) {
		return func(t *testing.T) {
			tg1 := tag.New("w:mno")
			ts1 := NewWithoutDuplicates([]tag.Tag{tg1})
			tg2 := tag.New("x:abc")
			ts2 := NewWithoutDuplicates([]tag.Tag{tg2})
			tg3 := tag.New("y:def")
			tg4 := tag.New("z:jkl")
			ts3 := NewWithoutDuplicates([]tag.Tag{tg3, tg4})

			expHash := hashOf("w:mno", "x:abc", "y:def", "z:jkl")

			// hash should be commutative and associative, so try a bunch
			// of combinations
			assert.Equal(t, expHash, union(ts1, union(ts2, ts3)).Hash())
			assert.Equal(t, expHash, union(ts1, union(ts3, ts2)).Hash())
			assert.Equal(t, expHash, union(union(ts2, ts3), ts1).Hash())
			assert.Equal(t, expHash, union(union(ts3, ts2), ts1).Hash())
			assert.Equal(t, expHash, union(ts3, union(ts1, ts2)).Hash())
			assert.Equal(t, expHash, union(ts3, union(ts2, ts1)).Hash())
			assert.Equal(t, expHash, union(union(ts1, ts2), ts3).Hash())
			assert.Equal(t, expHash, union(union(ts2, ts1), ts3).Hash())
		}
	}
	t.Run("Union", test(Union))
	t.Run("DisjointUnion", test(DisjointUnion))
}

// choose a random slice of vals.  It might be empty!
func chooseSubslice(r *rand.Rand, vals []byte) []byte {
	a := r.Intn(len(vals))
	b := r.Intn(len(vals)-a) + a
	return vals[a:b]
}

func bytesToTagSet(bytes []byte) *TagSet {
	tags := []tag.Tag{}
	for _, b := range bytes {
		tags = append(tags, tag.NewFromBytes([]byte{b}))
	}
	return New(tags)
}

func TestUnionOverlappingHashes(t *testing.T) {
	test := func(ts1 *TagSet, ts2 *TagSet, ts3 *TagSet, expHash uint64) func(*testing.T) {
		return func(t *testing.T) {
			// hash should be commutative and associative, so try a bunch
			// of combinations
			assert.Equal(t, expHash, Union(ts1, Union(ts2, ts3)).Hash())
			assert.Equal(t, expHash, Union(ts1, Union(ts3, ts2)).Hash())
			assert.Equal(t, expHash, Union(Union(ts2, ts3), ts1).Hash())
			assert.Equal(t, expHash, Union(Union(ts3, ts2), ts1).Hash())
			assert.Equal(t, expHash, Union(ts3, Union(ts1, ts2)).Hash())
			assert.Equal(t, expHash, Union(ts3, Union(ts2, ts1)).Hash())
			assert.Equal(t, expHash, Union(Union(ts1, ts2), ts3).Hash())
			assert.Equal(t, expHash, Union(Union(ts2, ts1), ts3).Hash())
		}
	}

	r := rand.New(rand.NewSource(13))
	letters := []byte("abcdefghijklmnopqrstuvwxyz")
	for i := 0; i < 100; i++ {
		slice1 := chooseSubslice(r, letters)
		ts1 := bytesToTagSet(slice1)
		slice2 := chooseSubslice(r, letters)
		ts2 := bytesToTagSet(slice2)
		slice3 := chooseSubslice(r, letters)
		ts3 := bytesToTagSet(slice3)

		// do a manual union of those slices
		seen := map[byte]struct{}{}
		for _, b := range slice1 {
			seen[b] = struct{}{}
		}
		for _, b := range slice2 {
			seen[b] = struct{}{}
		}
		for _, b := range slice3 {
			seen[b] = struct{}{}
		}

		// and compute the hash of that union
		var h uint64
		for b := range seen {
			h ^= murmur3.Sum64([]byte{b})
		}

		t.Run(fmt.Sprintf("%d: %s %s %s", i, slice1, slice2, slice3), test(ts1, ts2, ts3, h))
	}
}
