package main

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/djmitche/tagset/ident"
	"github.com/djmitche/tagset/loadgen"
	"github.com/djmitche/tagset/tagset"
)

const dataSize = 10

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

// global places for benchmarks to write, to avoid optimization
var HashH uint64
var HashL uint64
var bySize = map[int]loadgen.TagLineGenerator{}

func getTLG(size int) loadgen.TagLineGenerator {
	if cached, ok := bySize[size]; ok {
		return cached
	}

	tlg := loadgen.DSDTagLineGenerator()

	// lastly, create a preflight generator and preflight it so that
	// we do not measure tag-line generation time
	pf := loadgen.NewPreflightTagLineGenerator(size, tlg)
	pf.Preflight()

	bySize[size] = pf
	return pf
}

func benchmarkParsing(size int, b *testing.B) {
	tlg := getTLG(size)
	hostnames := loadgen.NewHostnameTagGenerator().GetTags()
	foundry := ident.NewInternFoundry()

	b.ResetTimer()

	global := tagset.NewWithoutDuplicates([]ident.Ident{
		foundry.Ident([]byte("region:antarctic")),
		foundry.Ident([]byte("epoch:holocene")),
	})
	for i := 0; i < b.N; i++ {
		common := tagset.DisjointUnion(global, tagset.NewWithoutDuplicates([]ident.Ident{
			foundry.Ident(<-hostnames),
		}))
		for line := range tlg.GetLines() {
			ts := tagset.Union(tagset.Parse(foundry, line), common)
			HashH ^= ts.HashH()
			HashL ^= ts.HashL()
		}
	}
}

func BenchmarkParsing1000(b *testing.B)   { benchmarkParsing(1000, b) }
func BenchmarkParsing10000(b *testing.B)  { benchmarkParsing(10000, b) }
func BenchmarkParsing100000(b *testing.B) { benchmarkParsing(100000, b) }
