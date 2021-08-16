package ident

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSorting(t *testing.T) {
	idents := []Ident{
		makeIdent("abc"),
		makeIdent("123"),
		makeIdent("xyz"),
		makeIdent("jkl"),
	}
	Sort(idents)

	require.True(t, idents[0].HashH() < idents[1].HashH())
	require.True(t, idents[1].HashH() < idents[2].HashH())
	require.True(t, idents[2].HashH() < idents[3].HashH())
}

func TestContains(t *testing.T) {
	idents := []Ident{
		makeIdent("abc"),
		makeIdent("123"),
		makeIdent("xyz"),
		makeIdent("jkl"),
	}
	Sort(idents)

	require.Equal(t, false, Contains(idents, makeIdent("XXX")))
	require.Equal(t, true, Contains(idents, idents[0]))
	require.Equal(t, true, Contains(idents, idents[1]))
	require.Equal(t, true, Contains(idents, idents[2]))
	require.Equal(t, true, Contains(idents, idents[3]))
	require.Equal(t, true, Contains(idents, makeIdent("abc")))
	require.Equal(t, true, Contains(idents, makeIdent("123")))
	require.Equal(t, true, Contains(idents, makeIdent("xyz")))
	require.Equal(t, true, Contains(idents, makeIdent("jkl")))
}

func TestContainsLetters(t *testing.T) {
	letters := strings.Split("g,h,i,j,k,l,m,n,o,p,q,r,s,t,u,v,w,x,y", ",")
	idents := []Ident{}
	for _, l := range letters {
		idents = append(idents, makeIdent(l))
	}

	Sort(idents)

	require.Equal(t, true, Contains(idents, makeIdent("m")))
	require.Equal(t, true, Contains(idents, makeIdent("n")))
	require.Equal(t, true, Contains(idents, makeIdent("o")))
	require.Equal(t, true, Contains(idents, makeIdent("p")))
	require.Equal(t, true, Contains(idents, makeIdent("q")))
}
