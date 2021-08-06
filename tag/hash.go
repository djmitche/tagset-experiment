package tag

import "github.com/twmb/murmur3"

const hashSize = 16

func hashTag(t []byte) (uint64, uint64) {
	return murmur3.Sum128(t)
}
