package ident

// A NullFoundry simply creates a new identifier for each call to Ident.  This
// can be used for tests or for infinite-cardinality identifiers (where each
// will only be seen once)
type NullFoundry struct{}

func NewNullFoundry() *NullFoundry {
	return &NullFoundry{}
}

func (f *NullFoundry) Ident(ident []byte) Ident {
	hashH, hashL := hashIdent(ident)
	return newIdent(ident, hashH, hashL)
}

func (f *NullFoundry) Get(hashH, hashL uint64) Ident {
	return nil
}
