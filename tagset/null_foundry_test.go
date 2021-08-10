package tagset

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/djmitche/tagset/ident"
	"github.com/stretchr/testify/suite"
	"github.com/twmb/murmur3"
)

type NullFoundrySuite struct {
	suite.Suite
	f *NullFoundry
}

func TestNullFoundry(t *testing.T) {
	suite.Run(t, &NullFoundrySuite{
		f: NewNullFoundry(),
	})
}

func (s *NullFoundrySuite) TestEmptyHash() {
	tg := idFoundry.Ident([]byte("x:abc"))
	ts := s.f.NewWithoutDuplicates([]ident.Ident{tg})
	expH, expL := hashOf("x:abc")
	gotH, gotL := ts.Hash()
	s.Equal(expH, gotH)
	s.Equal(expL, gotL)
}

func (s *NullFoundrySuite) TestSingleTagHash() {
	tg := idFoundry.Ident([]byte("x:abc"))
	ts := s.f.NewWithoutDuplicates([]ident.Ident{tg})
	expH, expL := hashOf("x:abc")
	gotH, gotL := ts.Hash()
	s.Equal(expH, gotH)
	s.Equal(expL, gotL)
}

func (s *NullFoundrySuite) TestParseEmpty() {
	ts := s.f.Parse(idFoundry, []byte{})
	s.Equal(uint64(0), ts.HashL())
	s.Equal(uint64(0), ts.HashH())
	s.Equal([]byte{}, ts.Serialization())
}

func (s *NullFoundrySuite) TestParseSingle() {
	ts := s.f.Parse(idFoundry, []byte("abc:def"))
	expH, expL := hashOf("abc:def")
	gotH, gotL := ts.Hash()
	s.Equal(expH, gotH)
	s.Equal(expL, gotL)
	s.Equal([]byte("abc:def"), ts.Serialization())
}

func (s *NullFoundrySuite) TestParseMulti() {
	ts := s.f.Parse(idFoundry, []byte("a,b,c"))
	expH, expL := hashOf("a", "b", "c")
	gotH, gotL := ts.Hash()
	s.Equal(expH, gotH)
	s.Equal(expL, gotL)
	// NOTE: it's not part of the API that this serialization has the same
	// order as the input, but in the current implementation that's the case.
	s.Equal([]byte("a,b,c"), ts.Serialization())
}

func (s *NullFoundrySuite) TestParseMultiDupes() {
	ts := s.f.Parse(idFoundry, []byte("a,b,a,b,c,c"))
	expH, expL := hashOf("a", "b", "c")
	gotH, gotL := ts.Hash()
	s.Equal(expH, gotH)
	s.Equal(expL, gotL)
	// NOTE: it's not part of the API that this serialization has the same
	// order as the input, but in the current implementation that's the case.
	s.Equal([]byte("a,b,c"), ts.Serialization())
}

func (s *NullFoundrySuite) TestFromBytes() {
	tg1 := idFoundry.Ident([]byte("x:abc"))
	tg2 := idFoundry.Ident([]byte("y:def"))
	ts := s.f.NewWithoutDuplicates([]ident.Ident{tg1, tg2})
	expH, expL := hashOf("x:abc", "y:def")
	gotH, gotL := ts.Hash()
	s.Equal(expH, gotH)
	s.Equal(expL, gotL)
}

func (s *NullFoundrySuite) TestTwoTagHash() {
	tg1 := idFoundry.Ident([]byte("x:abc"))
	tg2 := idFoundry.Ident([]byte("y:def"))

	expH, expL := hashOf("x:abc", "y:def")

	// hash should be the same regardless of order
	ts12 := s.f.NewWithoutDuplicates([]ident.Ident{tg1, tg2})
	got12H, got12L := ts12.Hash()
	s.Equal(expH, got12H)
	s.Equal(expL, got12L)

	ts21 := s.f.NewWithoutDuplicates([]ident.Ident{tg2, tg1})
	got21H, got21L := ts21.Hash()
	s.Equal(expH, got21H)
	s.Equal(expL, got21L)
}

func (s *NullFoundrySuite) TestSimpleDisjointUnions() {
	tg1 := idFoundry.Ident([]byte("w:mno"))
	ts1 := s.f.NewWithoutDuplicates([]ident.Ident{tg1})
	tg2 := idFoundry.Ident([]byte("x:abc"))
	ts2 := s.f.NewWithoutDuplicates([]ident.Ident{tg2})

	expH, expL := hashOf("w:mno", "x:abc")

	u := s.f.Union(ts1, ts2)
	gotH, gotL := u.Hash()

	s.Equal(expH, gotH)
	s.Equal(expL, gotL)
}

func (s *NullFoundrySuite) TestDisjointUnions() {
	test := func(union func(t1 *TagSet, t2 *TagSet) *TagSet) func() {
		return func() {
			tg1 := idFoundry.Ident([]byte("w:mno"))
			ts1 := s.f.NewWithoutDuplicates([]ident.Ident{tg1})
			tg2 := idFoundry.Ident([]byte("x:abc"))
			ts2 := s.f.NewWithoutDuplicates([]ident.Ident{tg2})
			tg3 := idFoundry.Ident([]byte("y:def"))
			tg4 := idFoundry.Ident([]byte("z:jkl"))
			ts3 := s.f.NewWithoutDuplicates([]ident.Ident{tg3, tg4})

			expH, expL := hashOf("w:mno", "x:abc", "y:def", "z:jkl")

			// hash should be commutative and associative, so try a bunch
			// of combinations
			check := func(unionedTs *TagSet) {
				s.Equal(expH, unionedTs.HashH(), "H")
				s.Equal(expL, unionedTs.HashL(), "L")
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
	s.Run("Union",
		test(func(t1 *TagSet, t2 *TagSet) *TagSet {
			return s.f.Union(t1, t2)
		}))
	s.Run("DisjointUnion",
		test(func(t1 *TagSet, t2 *TagSet) *TagSet {
			return s.f.DisjointUnion(t1, t2)
		}))
}

func (s *NullFoundrySuite) TestUnionOverlappingHashes() {
	test := func(ts1 *TagSet, ts2 *TagSet, ts3 *TagSet, expH, expL uint64) func() {
		return func() {
			check := func(unionedTs *TagSet) {
				s.Equal(expH, unionedTs.HashH())
				s.Equal(expL, unionedTs.HashL())
			}

			// hash should be commutative and associative, so try a bunch
			// of combinations
			check(s.f.Union(ts1, s.f.Union(ts2, ts3)))
			check(s.f.Union(ts1, s.f.Union(ts3, ts2)))
			check(s.f.Union(s.f.Union(ts2, ts3), ts1))
			check(s.f.Union(s.f.Union(ts3, ts2), ts1))
			check(s.f.Union(ts3, s.f.Union(ts1, ts2)))
			check(s.f.Union(ts3, s.f.Union(ts2, ts1)))
			check(s.f.Union(s.f.Union(ts1, ts2), ts3))
			check(s.f.Union(s.f.Union(ts2, ts1), ts3))
		}
	}

	r := rand.New(rand.NewSource(13))

	// choose a random slice of vals.  It might be empty!
	chooseSubslice := func(vals []byte) []byte {
		a := r.Intn(len(vals))
		b := r.Intn(len(vals)-a) + a
		return vals[a:b]
	}

	bytesToTagSet := func(bytes []byte) *TagSet {
		tags := []ident.Ident{}
		for _, b := range bytes {
			tags = append(tags, idFoundry.Ident([]byte{b}))
		}
		return s.f.NewWithDuplicates(tags)
	}

	letters := []byte("abcdefghijklmnopqrstuvwxyz")
	for i := 0; i < 100; i++ {
		slice1 := chooseSubslice(letters)
		ts1 := bytesToTagSet(slice1)
		slice2 := chooseSubslice(letters)
		ts2 := bytesToTagSet(slice2)
		slice3 := chooseSubslice(letters)
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

		s.Run(fmt.Sprintf("%d: %s %s %s", i, slice1, slice2, slice3), test(ts1, ts2, ts3, expH, expL))
	}
}
