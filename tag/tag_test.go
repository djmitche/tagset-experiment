package tag

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTagHash(t *testing.T) {
	tg := New("x:abc")
	expH, expL := hashTag([]byte("x:abc"))
	require.Equal(t, tg.HashH(), expH)
	require.Equal(t, tg.HashL(), expL)
	require.Equal(t, []byte("x:abc"), tg.Bytes())
}

func BenchmarkSameTagCreation(b *testing.B) {
	// use a different tag on each run
	tagBytes := fmt.Sprintf("tag:%d", b.N)
	for i := 0; i < b.N; i++ {
		New(tagBytes)
	}
}
