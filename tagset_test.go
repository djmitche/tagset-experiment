package tagset

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSingleTagHash(t *testing.T) {
	tg := NewTag("x:abc")
	ts := NewTagSetWithoutDuplicates([]*Tag{tg})
	require.Equal(t, hash("x:abc"), ts.Hash())
}

func TestFromBytes(t *testing.T) {
	tg1 := NewTag("x:abc")
	tg2 := NewTagFromBytes([]byte("y:def"))
	ts := NewTagSetWithoutDuplicates([]*Tag{tg1, tg2})
	require.Equal(t, hash("x:abc")^hash("y:def"), ts.Hash())
}

func TestTwoTagHash(t *testing.T) {
	tg1 := NewTag("x:abc")
	tg2 := NewTag("y:def")

	expHash := hash("x:abc") ^ hash("y:def")

	// hash should be the same regardless of order
	ts12 := NewTagSetWithoutDuplicates([]*Tag{tg1, tg2})
	require.Equal(t, expHash, ts12.Hash())

	ts21 := NewTagSetWithoutDuplicates([]*Tag{tg2, tg1})
	require.Equal(t, expHash, ts21.Hash())
}

func TestDisjointUnions(t *testing.T) {
	test := func(union func(t1 *TagSet, t2 *TagSet) *TagSet) func(*testing.T) {
		return func(t *testing.T) {
			tg1 := NewTag("x:abc")
			ts1 := NewTagSetWithoutDuplicates([]*Tag{tg1})
			tg2 := NewTag("y:def")
			ts2 := NewTagSetWithoutDuplicates([]*Tag{tg2})

			expHash := hash("x:abc") ^ hash("y:def")

			// hash should be the same regardless of order
			require.Equal(t, expHash, union(ts1, ts2).Hash())
			require.Equal(t, expHash, union(ts2, ts1).Hash())
		}
	}
	t.Run("Union", test(Union))
	t.Run("DisjointUnion", test(DisjointUnion))
}

func TestUnionOverlappingHashes(t *testing.T) {
	tg := NewTag("common")
	tg1 := NewTag("x:abc")
	ts1 := NewTagSetWithoutDuplicates([]*Tag{tg, tg1})
	tg2 := NewTag("y:def")
	ts2 := NewTagSetWithoutDuplicates([]*Tag{tg, tg2})

	expHash := hash("common") ^ hash("x:abc") ^ hash("y:def")

	// hash should be the same regardless of order
	require.Equal(t, expHash, Union(ts1, ts2).Hash())
	require.Equal(t, expHash, Union(ts2, ts1).Hash())
}
