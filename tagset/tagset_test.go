package tagset

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/djmitche/tagset/ident"
	"github.com/stretchr/testify/assert"
	"github.com/twmb/murmur3"
)

var foundry = ident.NewInternFoundry()

func hashOf(tags ...string) (uint64, uint64) {
	var hh, hl uint64
	for _, t := range tags {
		thh, thl := murmur3.Sum128([]byte(t))
		hh ^= thh
		hl ^= thl
	}
	return hh, hl
}

func TestSingleTagHash(t *testing.T) {
	tg := foundry.Ident([]byte("x:abc"))
	ts := NewWithoutDuplicates([]ident.Ident{tg})
	expH, expL := hashOf("x:abc")
	gotH, gotL := ts.Hash()
	assert.Equal(t, expH, gotH)
	assert.Equal(t, expL, gotL)
}

func TestParseEmpty(t *testing.T) {
	ts := Parse(foundry, []byte{})
	assert.Equal(t, uint64(0), ts.HashL())
	assert.Equal(t, uint64(0), ts.HashH())
	assert.Equal(t, []byte{}, ts.Serialization())
}

func TestParseSingle(t *testing.T) {
	ts := Parse(foundry, []byte("abc:def"))
	expH, expL := hashOf("abc:def")
	gotH, gotL := ts.Hash()
	assert.Equal(t, expH, gotH)
	assert.Equal(t, expL, gotL)
	assert.Equal(t, []byte("abc:def"), ts.Serialization())
}

func TestParseMulti(t *testing.T) {
	ts := Parse(foundry, []byte("a,b,c"))
	expH, expL := hashOf("a", "b", "c")
	gotH, gotL := ts.Hash()
	assert.Equal(t, expH, gotH)
	assert.Equal(t, expL, gotL)
	// NOTE: it's not part of the API that this serialization has the same
	// order as the input, but in the current implementation that's the case.
	assert.Equal(t, []byte("a,b,c"), ts.Serialization())
}

func TestParseMultiDupes(t *testing.T) {
	ts := Parse(foundry, []byte("a,b,a,b,c,c"))
	expH, expL := hashOf("a", "b", "c")
	gotH, gotL := ts.Hash()
	assert.Equal(t, expH, gotH)
	assert.Equal(t, expL, gotL)
	// NOTE: it's not part of the API that this serialization has the same
	// order as the input, but in the current implementation that's the case.
	assert.Equal(t, []byte("a,b,c"), ts.Serialization())
}

func TestFromBytes(t *testing.T) {
	tg1 := foundry.Ident([]byte("x:abc"))
	tg2 := foundry.Ident([]byte("y:def"))
	ts := NewWithoutDuplicates([]ident.Ident{tg1, tg2})
	expH, expL := hashOf("x:abc", "y:def")
	gotH, gotL := ts.Hash()
	assert.Equal(t, expH, gotH)
	assert.Equal(t, expL, gotL)
}

func TestTwoTagHash(t *testing.T) {
	tg1 := foundry.Ident([]byte("x:abc"))
	tg2 := foundry.Ident([]byte("y:def"))

	expH, expL := hashOf("x:abc", "y:def")

	// hash should be the same regardless of order
	ts12 := NewWithoutDuplicates([]ident.Ident{tg1, tg2})
	got12H, got12L := ts12.Hash()
	assert.Equal(t, expH, got12H)
	assert.Equal(t, expL, got12L)

	ts21 := NewWithoutDuplicates([]ident.Ident{tg2, tg1})
	got21H, got21L := ts21.Hash()
	assert.Equal(t, expH, got21H)
	assert.Equal(t, expL, got21L)
}

func TestSimpleDisjointUnions(t *testing.T) {
	tg1 := foundry.Ident([]byte("w:mno"))
	ts1 := NewWithoutDuplicates([]ident.Ident{tg1})
	tg2 := foundry.Ident([]byte("x:abc"))
	ts2 := NewWithoutDuplicates([]ident.Ident{tg2})

	expH, expL := hashOf("w:mno", "x:abc")

	u := Union(ts1, ts2)
	gotH, gotL := u.Hash()

	assert.Equal(t, expH, gotH)
	assert.Equal(t, expL, gotL)
}

func TestDisjointUnions(t *testing.T) {
	test := func(union func(t1 *TagSet, t2 *TagSet) *TagSet) func(*testing.T) {
		return func(t *testing.T) {
			tg1 := foundry.Ident([]byte("w:mno"))
			ts1 := NewWithoutDuplicates([]ident.Ident{tg1})
			tg2 := foundry.Ident([]byte("x:abc"))
			ts2 := NewWithoutDuplicates([]ident.Ident{tg2})
			tg3 := foundry.Ident([]byte("y:def"))
			tg4 := foundry.Ident([]byte("z:jkl"))
			ts3 := NewWithoutDuplicates([]ident.Ident{tg3, tg4})

			expH, expL := hashOf("w:mno", "x:abc", "y:def", "z:jkl")

			// hash should be commutative and associative, so try a bunch
			// of combinations
			check := func(unionedTs *TagSet) {
				assert.Equal(t, expH, unionedTs.HashH(), "H")
				assert.Equal(t, expL, unionedTs.HashL(), "L")
			}
			check(union(ts1, union(ts2, ts3)))
			check(union(ts1, union(ts3, ts2)))
			check(union(union(ts2, ts3), ts1))
			check(union(union(ts3, ts2), ts1))
			check(union(ts3, union(ts1, ts2)))
			check(union(ts3, union(ts2, ts1)))
			check(union(union(ts1, ts2), ts3))
			check(union(union(ts2, ts1), ts3))
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
	tags := []ident.Ident{}
	for _, b := range bytes {
		tags = append(tags, foundry.Ident([]byte{b}))
	}
	return New(tags)
}

func TestUnionOverlappingHashes(t *testing.T) {
	test := func(ts1 *TagSet, ts2 *TagSet, ts3 *TagSet, expH, expL uint64) func(*testing.T) {
		return func(t *testing.T) {
			check := func(unionedTs *TagSet) {
				assert.Equal(t, expH, unionedTs.HashH())
				assert.Equal(t, expL, unionedTs.HashL())
			}

			// hash should be commutative and associative, so try a bunch
			// of combinations
			check(Union(ts1, Union(ts2, ts3)))
			check(Union(ts1, Union(ts3, ts2)))
			check(Union(Union(ts2, ts3), ts1))
			check(Union(Union(ts3, ts2), ts1))
			check(Union(ts3, Union(ts1, ts2)))
			check(Union(ts3, Union(ts2, ts1)))
			check(Union(Union(ts1, ts2), ts3))
			check(Union(Union(ts2, ts1), ts3))
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
		var expH, expL uint64
		for b := range seen {
			h, l := murmur3.Sum128([]byte{b})
			expH ^= h
			expL ^= l
		}

		t.Run(fmt.Sprintf("%d: %s %s %s", i, slice1, slice2, slice3), test(ts1, ts2, ts3, expH, expL))
	}
}
