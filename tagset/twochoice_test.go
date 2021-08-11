package tagset

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func fakeTagSetWithHash(hashH, hashL uint64) *TagSet {
	return &TagSet{
		hashH: hashH,
		hashL: hashL,
	}
}

func twoChoiceTestCount() uint64 {
	count := uint64(hashMask)
	if count > 1000 {
		// too big to fill, so just use "some"
		count = 1000
	} else {
		// hold back a bit from the maximum capacity of the table
		count = count - 10
	}
	return count
}

// Check that TwoChoice behaves like a regular old map
func TestTwoChoiceWellBehaved(t *testing.T) {
	count := twoChoiceTestCount()

	tbl := newTwoChoice(int(count))
	regularMap := map[string]*TagSet{}

	for i := uint64(0); i < count; i++ {
		h := rand.Uint64() & hashMask
		l := rand.Uint64() & hashMask
		k := fmt.Sprintf("%016x.%016x", h, l)

		if _, found := regularMap[k]; found {
			elt := tbl.get(h, l)
			require.NotNil(t, elt)
			require.Equal(t, h, elt.hashH)
			require.Equal(t, l, elt.hashL)
		} else {
			require.Nil(t, tbl.get(h, l))
		}

		ts := fakeTagSetWithHash(h, l)
		regularMap[k] = ts
		tbl.insert(h, l, ts)
	}

	// now try to find all of those elements again..
	for k, ts := range regularMap {
		h := ts.hashH
		l := ts.hashL
		got := tbl.get(h, l)
		require.Equal(t, ts, got, k)
	}
}

func TestTwoChoiceCollisions(t *testing.T) {
	count := twoChoiceTestCount()
	tbl := newTwoChoice(int(count))

	base := uint64(0x123456789)

	// insert (checking for nil), using hashes that are guaranteed
	// to generate lots of collisions
	for i := uint64(0); i < count; i++ {
		hashH := (base + i/2) & hashMask
		hashL := (base + i) & hashMask
		ts := fakeTagSetWithHash(hashH, hashL)

		require.Nil(t, tbl.get(hashH, hashL))
		tbl.insert(hashH, hashL, ts)
	}

	// insert again (should overwrite)
	for i := uint64(0); i < count; i++ {
		hashH := (base + i/2) & hashMask
		hashL := (base + i) & hashMask
		elt := tbl.get(hashH, hashL)
		require.Equal(t, hashH, elt.HashH())
		require.Equal(t, hashL, elt.HashL())

		ts := fakeTagSetWithHash(hashH, hashL)
		tbl.insert(hashH, hashL, ts)
	}

	// get results
	for i := uint64(0); i < count; i++ {
		hashH := (base + i/2) & hashMask
		hashL := (base + i) & hashMask

		elt := tbl.get(hashH, hashL)
		require.Equal(t, hashH, elt.HashH())
		require.Equal(t, hashL, elt.HashL())
	}
}

func BenchmarkTwoChoice(b *testing.B) {
	baseH := rand.Uint64()
	baseL := rand.Uint64()
	tbl := newTwoChoice(b.N)
	n := uint64(b.N)

	if n > hashMask {
		b.Skipf("don't benchmark with small hashMask")
	}

	b.ResetTimer()
	b.ReportAllocs()

	// we only store a single tagset, to avoid wasting time
	// on this allocation
	ts := fakeTagSetWithHash(baseH, baseL)

	for i := uint64(0); i < n; i++ {
		hashH := (baseH + i) & hashMask
		hashL := (baseL + i) & hashMask
		tbl.insert(hashH, hashL, ts)
	}
}
