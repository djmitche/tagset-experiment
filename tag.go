package tagset

// TODO: where to avoid allocations, e.g., when splitting a string of tags?

type Tag struct {
	t []byte
	// TODO: do we need to store this here?
	hash uint64
}

// NewTag creates a new tag from a given byte array.  The byte slice is no longer
// referenced after return from this function.
func NewTagFromBytes(t []byte) *Tag {
	clone := make([]byte, len(t), len(t))
	copy(clone, t)
	return &Tag{
		t:    clone,
		hash: hashTag(t),
	}
}

// NewTag creates a new tag from a given string, computing its hash in the
// process.
func NewTag(t string) *Tag {
	return NewTagFromBytes([]byte(t))
}
