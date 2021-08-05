package tag

import "github.com/twmb/murmur3"

const hashSize = 8

func hashTag(t []byte) uint64 {
	return murmur3.Sum64(t)
}
