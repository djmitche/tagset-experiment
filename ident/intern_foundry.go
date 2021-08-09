package ident

/* IMPLEMENTATION NOTES
 *
 * This foundry uses [2-choice hashing](https://en.wikipedia.org/wiki/2-choice_hashing) with
 * eviction on collision.  It uses the high- and low-order 64 bits of the hash as the two
 * distinct hashes.  It compares all 128 bits to determine a hit.  On a collision of either
 * the high or low bits during insertion, it overwrites the existing Ident.  This means that
 * in the case of a collision, some interned data may be duplicated, failing safe.
 *
 * Note that Murmur3 is _not_ cryptographically collision-resistant, so a
 * particularly abusive user of this foundry could, with moderate effort, cause
 * significant duplicated data.  For purposes of agent performance, this is not
 * an issue as users generally want the agent to perform well.
 */

// A InternFoundry caches identifiers forever, effectively acting like a
// string interner.
type InternFoundry struct {
	byHash map[uint64]Ident
}

func NewInternFoundry() *InternFoundry {
	return &InternFoundry{byHash: map[uint64]Ident{}}
}

func (f *InternFoundry) Ident(ident []byte) Ident {
	hashH, hashL := hashIdent(ident)
	existing := f.get(hashH, hashL)
	if existing != nil {
		return existing
	}

	rv := newIdent(ident, hashH, hashL)
	f.insert(hashH, hashL, rv)

	return rv
}

func (f *InternFoundry) get(hashH, hashL uint64) Ident {
	var hit Ident

	hit = f.byHash[hashH]
	if hit != nil && hit.HashL() == hashL && hit.HashH() == hashH {
		return hit
	}

	hit = f.byHash[hashL]
	if hit != nil && hit.HashL() == hashL && hit.HashH() == hashH {
		return hit
	}

	return nil
}

func (f *InternFoundry) insert(hashH, hashL uint64, ident Ident) {
	f.byHash[hashH] = ident
	f.byHash[hashL] = ident
}
