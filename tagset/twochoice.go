package tagset

// A twoChoice implements a hash table indexed by a 128-bit value represented
// as a high and low uint64.  Internally, it uses Go's `map` and the two-choice
// hashing technique to keep lookups O(log N).  Collisions are handled with open
// chaining on the high uint64, but are exceedingly rare.  See `hashMask` for help
// testing this condition.
type twoChoice map[uint64]*TagSet

// Create a new twoChoice map, optionally with a capacity.
func newTwoChoice(capacity ...int) twoChoice {
	var cap int
	if len(capacity) > 0 {
		cap = capacity[0]
	} else {
		cap = 0
	}

	// allocate a map with double the capacity, since we insert two choices.
	return make(twoChoice, cap*2)
}

// Get a TagSet, if it is present in the hash table.
func (tbl twoChoice) get(hashH, hashL uint64) *TagSet {
	hashH &= hashMask
	hashL &= hashMask

	// first choice..
	elt, foundH := tbl[hashH]
	if foundH && (elt.HashL()&hashMask) == hashL {
		return elt
	}

	// second choice..
	elt, foundL := tbl[hashL]
	if foundL && (elt.HashH()&hashMask) == hashH {
		return elt
	}

	// open chaining
	if foundH {
		for {
			hashH = (hashH + 1) & hashMask
			elt, found := tbl[hashH]
			if found {
				if (elt.HashL() & hashMask) == hashL {
					return elt
				}
				continue
			} else {
				// nothing in this slot -> not found
				return nil
			}
		}
	}

	return nil
}

// Insert a TagSet, overwriting element any already present with the same hash.
func (tbl twoChoice) insert(hashH, hashL uint64, newElt *TagSet) {
	hashH &= hashMask
	hashL &= hashMask

	// first choice..
	elt, found := tbl[hashH]
	if !found || (elt.HashL()&hashMask) == hashL {
		tbl[hashH] = newElt
		return
	}

	// second choice..
	elt, found = tbl[hashL]
	if !found || (elt.HashH()&hashMask) == hashH {
		tbl[hashL] = newElt
		return
	}

	// open chaining
	origHashH := hashH & hashMask
	for {
		hashH = (hashH + 1) & hashMask
		if hashH == origHashH {
			// we've scanned the full table; this would require 2**128 entries without
			// hashmask, so it's safe to assume it will never happen.
			panic("twochoice table full")
		}

		elt, found = tbl[hashH]
		if found {
			if (elt.HashL() & hashMask) == hashL {
				tbl[hashH] = newElt
				return
			}
			continue
		} else {
			// nothing in this slot -> not found, so put it here
			tbl[hashH] = newElt
			return
		}
	}
}
