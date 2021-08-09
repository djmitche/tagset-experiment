package ident

// A Foundry produces identifiers.  In general, foundries are not threadsafe,
// except where explicitly specified.
type Foundry interface {
	// Ident returns an Ident for the given byte slice.  The byte slice is
	// not maintained, and the caller may reuse it.
	Ident([]byte) Ident
}
