package tagset

// A twoChoice implements a hash table indexed by a 128-bit value represented
// as a high and low uint64.  Internally, it uses Go's `map` and the two-choice
// hashing technique to keep lookups O(log N).  Collisions are handled with open
// chaining on the high uint64, but are exceedingly rare.  See `hashMask` for help
// testing this condition.
type twoChoice map[uint64]twoChoiceElt

type twoChoiceElt struct {
	hashH, hashL uint64
	ts           *TagSet
}

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
	if foundH && (elt.hashL&hashMask) == hashL && (elt.hashH&hashMask) == hashH {
		return elt.ts
	}

	// second choice..
	elt, foundL := tbl[hashL]
	if foundL && (elt.hashH&hashMask) == hashH && (elt.hashL&hashMask) == hashL {
		return elt.ts
	}

	// open chaining, when the hashH element existed but hashL didn't match
	if foundH {
		chainH := hashH
		for {
			chainH = (chainH + 1) & hashMask
			if chainH == hashH {
				// we've scanned the full table; this would require 2**128 entries with
				// a full-sized hashMask, so it's safe to assume it will never happen.
				panic("twochoice table full")
			}

			elt, found := tbl[chainH]
			if found {
				if (elt.hashL&hashMask) == hashL && (elt.hashH&hashMask) == hashH {
					return elt.ts
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

	// hash at the first choice..
	elt, collisH := tbl[hashH]
	if !collisH || ((elt.hashL&hashMask) == hashL && (elt.hashH&hashMask) == hashH) {
		tbl[hashH] = twoChoiceElt{hashH: hashH, hashL: hashL, ts: newElt}
		return
	}

	// failing that, at the second choice..
	elt, collisL := tbl[hashL]
	if !collisL || ((elt.hashH&hashMask) == hashH && (elt.hashL&hashMask) == hashL) {
		tbl[hashL] = twoChoiceElt{hashH: hashH, hashL: hashL, ts: newElt}
		return
	}

	// if both of those collided, resort to open chaining
	chainH := hashH
	for {
		chainH = (chainH + 1) & hashMask
		if chainH == hashH {
			// we've scanned the full table; this would require 2**128 entries with
			// a full-sized hashMask, so it's safe to assume it will never happen.
			panic("twochoice table full")
		}

		elt, found := tbl[chainH]
		if found {
			if (elt.hashL&hashMask) == hashL && (elt.hashH&hashMask) == hashH {
				tbl[chainH] = twoChoiceElt{hashH: hashH, hashL: hashL, ts: newElt}
				return
			}
			continue
		} else {
			// nothing in this slot -> not found, so put it here
			tbl[chainH] = twoChoiceElt{hashH: hashH, hashL: hashL, ts: newElt}
			return
		}
	}
}
