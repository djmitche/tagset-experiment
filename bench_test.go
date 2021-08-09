package main

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/djmitche/tagset/ident"
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
var bySize = map[int]tagLineGenerator{}

func getTLG(size int) tagLineGenerator {
	if cached, ok := bySize[size]; ok {
		return cached
	}

	// almost every line has an environment
	envs := newLowCardinalityTagGenerator([][]byte{
		[]byte("env:staging"),
		[]byte("env:prod"),
		[]byte("env:dev"),
		[]byte("env:qa"),
		[]byte("env:lab"),
		[]byte("env:bench"),
		[]byte("env:rainforest"),
	})

	// some random tags that might or might not appear on a tag line
	tags := newLowCardinalityTagGenerator([][]byte{
		[]byte("library:snorkels"),
		[]byte("library:bunny"),
		[]byte("library:bananas"),
		[]byte("library:pamplemousse"),
		[]byte("abtest:123"),
		[]byte("abtest:992"),
		[]byte("abtest:16"),
		[]byte("shard:x"),
		[]byte("shard:y"),
		[]byte("shard:z"),
	})

	var lowcard tagLineGenerator = newRandomTagLineGenerator([]randomTagLineOptions{{
		tg:      envs,
		prob:    99,
		repeats: 1,
	}, {
		tg:      tags,
		prob:    25,
		repeats: 3,
	}})

	// duplicate those, with an active set of 100 tag lines each
	// repeated up to 20 times
	lowcard = newDuplicateTagLineGenerator(lowcard, 100, 20)

	spans := newRandomTagLineGenerator([]randomTagLineOptions{{
		tg:      envs,
		prob:    99,
		repeats: 1,
	}, {
		tg:      newHighCardinalityTagGenerator("span", size/10),
		prob:    80,
		repeats: 1,
	}})

	var tlg tagLineGenerator = newMultiplexingTagLineGenerator([]tagLineGenerator{
		lowcard,
		spans,
	})

	// lastly, create a preflight generator and preflight it so that
	// we do not measure tag-line generation time
	pf := newPreflightTagLineGenerator(size, tlg)
	pf.preflight()

	bySize[size] = pf
	return pf
}

func benchmarkParsing(size int, b *testing.B) {
	tlg := getTLG(size)
	hostnames := newHostnameTagGenerator().getTags()
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
		for line := range tlg.getLines() {
			ts := tagset.Union(tagset.Parse(foundry, line), common)
			HashH ^= ts.HashH()
			HashL ^= ts.HashL()
		}
	}
}

func BenchmarkParsing1000(b *testing.B)   { benchmarkParsing(1000, b) }
func BenchmarkParsing10000(b *testing.B)  { benchmarkParsing(10000, b) }
func BenchmarkParsing100000(b *testing.B) { benchmarkParsing(100000, b) }
