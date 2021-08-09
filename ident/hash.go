package ident

import "github.com/twmb/murmur3"

const hashSize = 16

func hashIdent(t []byte) (uint64, uint64) {
	return murmur3.Sum128(t)
}
