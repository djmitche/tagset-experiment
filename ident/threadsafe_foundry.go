package ident

import "sync"

// A ThreadsafeFoundry wraps another Foundry and applies locking to allow
// concurrent access from multiple goroutines.
type ThreadsafeFoundry struct {
	sync.Mutex
	inner Foundry
}

func NewThreadsafeFoundry(inner Foundry) *ThreadsafeFoundry {
	return &ThreadsafeFoundry{
		sync.Mutex{},
		inner,
	}
}

func (f *ThreadsafeFoundry) Ident(ident []byte) Ident {
	f.Lock()
	rv := f.inner.Ident(ident)
	f.Unlock()
	return rv
}
