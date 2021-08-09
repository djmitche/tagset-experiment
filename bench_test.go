package main

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/djmitche/tagset/ident"
	"github.com/djmitche/tagset/tagset"
)

const dataSize = 1000000

var data [][]byte

func init() {
	// create a *lot* of tag sets, as `,`-separated strings

	lowCards := [][]byte{
		[]byte("app:foo"),
		[]byte("app:bar"),
		[]byte("app:bing"),
		[]byte("app:baz"),
		[]byte("env:prod"),
		[]byte("env:staging"),
		[]byte("env:dev"),
		[]byte("env:playground"),
		[]byte("planet:earth"),
	}
	lowCard := func() []byte {
		i := rand.Intn(len(lowCards))
		return lowCards[i]
	}

	midCard := func() []byte {
		i := rand.Intn(32768)
		return []byte(fmt.Sprintf("mid:%d", i))
	}

	highCard := func() []byte {
		i := rand.Uint64()
		return []byte(fmt.Sprintf("high:%#v", i))
	}

	for i := 0; i < dataSize; i++ {
		var n int

		// sometimes, repeat ourselves
		n = rand.Intn(dataSize)
		if n < i {
			data = append(data, data[n])
			continue
		}

		var datum []byte

		n = rand.Intn(20) + 1
		for j := 0; j < n; j++ {
			datum = append(datum, lowCard()...)
			datum = append(datum, byte(','))
		}
		n = rand.Intn(20)
		for j := 0; j < n; j++ {
			datum = append(datum, midCard()...)
			datum = append(datum, byte(','))
		}
		n = rand.Intn(20)
		for j := 0; j < n; j++ {
			datum = append(datum, highCard()...)
			datum = append(datum, byte(','))
		}

		// strip the final `,` from datum
		data = append(data, datum[:len(datum)-1])
	}
}

func benchmarkParsing(size int, b *testing.B) {
	foundry := ident.NewInternFoundry()
	global := tagset.NewWithoutDuplicates([]ident.Ident{foundry.Ident([]byte("planet:earth")), foundry.Ident([]byte("epoch:holocene"))})
	inc := len(data) / size
	for i := 0; i < b.N; i++ {
		common := tagset.DisjointUnion(global, tagset.NewWithoutDuplicates([]ident.Ident{foundry.Ident([]byte(("host:i-1029812")))}))
		for j := 0; j < len(data); j += inc {
			tagset.Union(tagset.Parse(foundry, data[i]), common)
		}
	}
}

func BenchmarkParsing100(b *testing.B) {
	benchmarkParsing(100, b)
}

func BenchmarkParsing10000(b *testing.B) {
	benchmarkParsing(10000, b)
}

func BenchmarkParsing1000000(b *testing.B) {
	benchmarkParsing(1000000, b)
}
