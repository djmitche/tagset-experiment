package tagset

import (
	"testing"

	"github.com/djmitche/tagset/ident"
	"github.com/stretchr/testify/assert"
)

func TestTagsetAccessors(t *testing.T) {
	tg := idFoundry.Ident([]byte("x:abc"))
	ts := &TagSet{
		size:          1,
		tags:          []ident.Ident{tg},
		hashH:         tg.HashH(),
		hashL:         tg.HashL(),
		serialization: []byte("x:abc"),
	}
	gotH, gotL := ts.Hash()
	assert.Equal(t, tg.HashH(), gotH)
	assert.Equal(t, tg.HashL(), gotL)
	gotH = ts.HashH()
	gotL = ts.HashL()
	assert.Equal(t, tg.HashH(), gotH)
	assert.Equal(t, tg.HashL(), gotL)
	ser := ts.Serialization()
	assert.Equal(t, []byte("x:abc"), ser)
}

// TODO: test `has`, `forEach` if they still exist
