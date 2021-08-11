package tagset

// (used in parsing)
var commaSeparator = []byte(",")

// A guess at tag size (16), to eliminate a few unnecessary reallocations of serializations
const avgTagSize = 16

// set this to 0xfff to test hash collisions.  The compiler optimizes this out
// when it is `^uint64(0)`.
const hashMask = ^uint64(0)
