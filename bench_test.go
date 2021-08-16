package main

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/djmitche/tagset/ident"
	"github.com/djmitche/tagset/loadgen"
	"github.com/djmitche/tagset/tagset"
	"github.com/stretchr/testify/require"
)

// global places for benchmarks to write, to avoid optimization
var Line []byte
var TS *tagset.TagSet

// Benchmark reading from a channel
func BenchmarkChannel(b *testing.B) {
	b.Skip("Here just for reference")
	c := make(chan ([]byte), 1024)
	go func() {
		buf := make([]byte, 0, 10)
		for i := b.N; i > 0; i-- {
			c <- buf
		}
		close(c)
	}()
	for b := range c {
		Line = b
	}
}

/* Things to know while interpreting benchmark data
 *
 * - channels themselves take about 100ns per item
 * - the DSD generator can produce about one item per 500ns
 */

// Benchmark the tag-line generator to get a baseline over which the
// parsing can be measured
func BenchmarkGenerator(b *testing.B) {
	b.Skip("Here just for reference")
	tlg := loadgen.NewCmdTagLineGenerator("hyper", b.N)
	lines := tlg.GetLines()

	b.ReportAllocs()
	b.ResetTimer()

	count := 0
	for line := range lines {
		count += 1
		Line = line
	}
	require.Equal(b, count, b.N)
}

var parsingNoteOnce sync.Once

func benchmarkParsing(b *testing.B, tsFoundry tagset.Foundry) {
	// operate at 1000x the benchmarks, because otherwise allocs/op rounds
	// to the nearest integer and loses precision
	n := 1000 * b.N
	parsingNoteOnce.Do(func() {
		b.Log("NOTE: for `Benchmark.*FoundryParsing`, one 'op' means parsing 1000 lines.")
	})

	const warmupCount = 1000
	tlg := loadgen.NewCmdTagLineGenerator("dsd", n+warmupCount)
	lines := tlg.GetLines()
	idFoundry := ident.NewInternFoundry()

	// warm up the parser first
	for i := 0; i < warmupCount; i++ {
		tsFoundry.Parse(idFoundry, <-lines)
	}

	b.ReportAllocs()
	b.ResetTimer()

	count := 0
	almostDone := n * 10 / 9
	if almostDone < 1000 {
		almostDone = 1000
	}
	for line := range lines {
		// double-check that we are not waiting on the producer near
		// the end of the benchmark run
		if count == almostDone {
			require.NotEqual(b,
				len(lines), 0,
				"tag-line producer is not keeping up after %d items", count)
		}

		count++

		TS = tsFoundry.Parse(idFoundry, line)
	}

	b.StopTimer()

	require.Equal(b, count, n)
}

func BenchmarkNullFoundryParsing(b *testing.B) { benchmarkParsing(b, tagset.NewNullFoundry()) }
func BenchmarkInternFoundryParsing(b *testing.B) {
	f := tagset.NewInternFoundry()
	benchmarkParsing(b, f)
	// report the percent of parses that missed the cache
	b.ReportMetric(float64(f.ParseMisses)*100/float64(f.Parses), "miss%")
}

func benchmarkUnion(b *testing.B, tsFoundry tagset.Foundry) {
	// operate at 1000x the benchmarks, because otherwise allocs/op rounds
	// to the nearest integer and loses precision
	n := 1000 * b.N
	parsingNoteOnce.Do(func() {
		b.Log("NOTE: for `Benchmark.*FoundryParsing`, one 'op' means parsing 1000 lines.")
	})

	// get ourselves a collection of tagsets
	const tagsetCount = 100
	tagsets := make([]*tagset.TagSet, 0, tagsetCount)
	tlg := loadgen.NewCmdTagLineGenerator("dsd", tagsetCount)
	lines := tlg.GetLines()
	idFoundry := ident.NewInternFoundry()
	for i := 0; i < tagsetCount; i++ {
		tagsets = append(tagsets, tsFoundry.Parse(idFoundry, <-lines))
	}

	test := func(count int) {
		for i := 0; i < count; i++ {
			TS = tsFoundry.Union(tagsets[rand.Intn(tagsetCount)], tagsets[rand.Intn(tagsetCount)])
		}
	}

	// warm things up a bit..
	test(1000)

	b.ReportAllocs()
	b.ResetTimer()

	test(n)
}

func BenchmarkNullFoundryUnion(b *testing.B)   { benchmarkUnion(b, tagset.NewNullFoundry()) }
func BenchmarkInternFoundryUnion(b *testing.B) { benchmarkUnion(b, tagset.NewInternFoundry()) }
