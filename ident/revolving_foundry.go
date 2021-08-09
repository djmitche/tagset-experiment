package ident

// A RevolvingFoundry contains multiple InternFoundries and rotates through
// them, allowing identifiers which are no longer used to be freed.  The effect
// is similar to a batched least-recently-used cache.
type RevolvingFoundry struct {
	rotateAfter int
	count       int
	inner       []*InternFoundry
}

// Create a RevolvingFoundry of the given size (number of InternFoundries) and
// rotating after the given number of accesses.
//
// The size should typically be very small (single digits, at least 2), as it
// impacts the time it takes to create a novel Ident.
//
// The `rotateAfter` parameter should be tuned so that `rotateAfter` times
// `size - 1` approximates the cardinality of the identifiers being created.
// For example, if the identifiers are hostnames and there are typically 10,000
// hosts active at any time, then `rotateAfter = 5000` and `size = 3` are
// good choices.
func NewRevolvingFoundry(size, rotateAfter int) *RevolvingFoundry {
	if size < 2 {
		panic("size must be at least 2")
	}
	inner := make([]*InternFoundry, size)
	for i, _ := range inner {
		inner[i] = NewInternFoundry()
	}

	return &RevolvingFoundry{
		rotateAfter: rotateAfter,
		count:       0,
		inner:       inner,
	}
}

func (f *RevolvingFoundry) Ident(ident []byte) Ident {
	f.count++
	hashH, hashL := hashIdent(ident)

	if f.count > f.rotateAfter {
		f.rotate()
		f.count = 0
	}

	// search through the inner foundries for an existing interned
	// value, moving it to the first foundry if found
	for i, inner := range f.inner {
		hit := inner.get(hashH, hashL)
		if hit != nil {
			// if this hit was not in the first inner foundry, add it there
			if i > 0 {
				f.inner[0].insert(hashH, hashL, hit)
			}
			return hit
		}
	}

	// not found, so add it to the first inner foundry
	rv := newIdent(ident, hashH, hashL)
	f.inner[0].insert(hashH, hashL, rv)

	return rv
}

// Insert a new InternFoundry at the beginning of the rotation, dropping the
// last foundry.
func (f *RevolvingFoundry) rotate() {
	newInner := make([]*InternFoundry, len(f.inner))
	newInner[0] = NewInternFoundry()
	copy(newInner[1:], f.inner[:len(f.inner)-1])
	f.inner = newInner
}
