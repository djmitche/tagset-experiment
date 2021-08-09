package ident

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRevolvingFoundry(t *testing.T) {
	f := NewRevolvingFoundry(3, 5)

	// fill one inner with a's
	var aIds []Ident
	for i := 0; i < 5; i++ {
		ident := f.Ident([]byte(fmt.Sprintf("a:%d", i)))
		aIds = append(aIds, ident)
	}

	// pull one of those forward (rotating in the process)
	f.Ident([]byte("a:3"))

	var bIds []Ident
	for i := 0; i < 5; i++ {
		ident := f.Ident([]byte(fmt.Sprintf("b:%d", i)))
		bIds = append(bIds, ident)
	}

	var cIds []Ident
	for i := 0; i < 5; i++ {
		ident := f.Ident([]byte(fmt.Sprintf("c:%d", i)))
		cIds = append(cIds, ident)
	}

	// that `a:3` should still be around, but the other a's not
	a3 := f.Ident([]byte("a:3"))
	assert.True(t, &a3[0] == &aIds[3][0])
	a4 := f.Ident([]byte("a:4"))
	assert.False(t, &a4[0] == &aIds[4][0])

	// and the b's should still be around
	b0 := f.Ident([]byte("b:0"))
	assert.True(t, &b0[0] == &bIds[0][0])
}
