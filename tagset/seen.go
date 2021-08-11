package tagset

// A seenTracker can be used to track seeing things by their 128-bit hashes
type seenTracker map[uint64][]uint64

// Track the given identifier as seen and return true if it had been seen
// before.
func (seen seenTracker) seen(hashH, hashL uint64) bool {
	// the map uses the high uint64 as a hash table index, with a linear search
	// used to find the low uint64 in the bucket.  Almost every bucket will be
	// one item long.  NOTE: test the bucketing by stripping bits from `hashH`
	// with `hashH &= 0x7`.

	hashBucket, found := seen[hashH]
	if found {
		found = false
		for _, existingL := range hashBucket {
			if existingL == hashL {
				found = true
				break
			}
		}
	} else {
		hashBucket = []uint64{}
	}
	if !found {
		seen[hashH] = append(hashBucket, hashL)
	}

	return found
}
