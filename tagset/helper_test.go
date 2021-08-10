package tagset

import (
	"github.com/djmitche/tagset/ident"
	"github.com/twmb/murmur3"
)

var idFoundry = ident.NewInternFoundry()

func hashOf(tags ...string) (uint64, uint64) {
	var hh, hl uint64
	for _, t := range tags {
		thh, thl := murmur3.Sum128([]byte(t))
		hh ^= thh
		hl ^= thl
	}
	return hh, hl
}
