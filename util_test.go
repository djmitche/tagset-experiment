package tagset

// convenience function to hash strings in tests
func hash(t string) uint64 {
	return hashTag([]byte(t))
}
