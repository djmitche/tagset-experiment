package tagset

import "github.com/twmb/murmur3"

func hashTag(t []byte) uint64 {
	return murmur3.Sum64(t)
}
