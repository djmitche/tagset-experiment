package ident

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestThreadsafeFoundry(t *testing.T) {
	f := NewThreadsafeFoundry(NewInternFoundry())

	var id1, id2 Ident
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		id1 = f.Ident([]byte("abc:def"))
		wg.Done()
	}()
	go func() {
		id2 = f.Ident([]byte("abc:def"))
		wg.Done()
	}()

	wg.Wait()

	// InternFoundry should have deduplicated these
	require.True(t, &id1[0] == &id2[0])
}
