package ident

import (
	"sort"
)

type identSlice []Ident

// Len implements a function in `sort.Interface` for slices of identifiers.
func (idents *identSlice) Len() int {
	return len(*idents)
}

// Less implements a function in `sort.Interface` for slices of identifiers.
func (idents *identSlice) Less(i, j int) bool {
	return (*idents)[i].Less((*idents)[j])
}

// Swap implements a function in `sort.Interface` for slices of identifiers.
func (idents *identSlice) Swap(i, j int) {
	(*idents)[i], (*idents)[j] = (*idents)[j], (*idents)[i]
}

// Sort sorts the given slice of identifiers by hash, using `sort.Sort`
func Sort(idents []Ident) {
	slice := identSlice(idents)
	sort.Sort(&slice)
}

// Search searches for the given identifier int the given _sorted_ slice of
// identifiers, using `sort.Search`, returning true if it was found.
func Contains(haystack []Ident, needle Ident) bool {
	n := len(haystack)
	i := sort.Search(n, func(i int) bool {
		return !haystack[i].Less(needle)
	})
	if i >= n {
		return false
	}
	return haystack[i].Equals(needle)
}
