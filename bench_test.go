package main

import (
	"testing"

	"github.com/djmitche/tagset/ident"
	"github.com/djmitche/tagset/loadgen"
	"github.com/djmitche/tagset/tagset"
	"github.com/stretchr/testify/require"
)

// global places for benchmarks to write, to avoid optimization
var HashH uint64
var HashL uint64
var Line []byte

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

func BenchmarkParsing(b *testing.B) {
	tlg := loadgen.NewCmdTagLineGenerator("dsd", b.N)
	lines := tlg.GetLines()
	idFoundry := ident.NewInternFoundry()
	tsFoundry := tagset.NewNullFoundry()

	b.ReportAllocs()
	b.ResetTimer()

	count := 0
	almostDone := b.N * 10 / 9
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

		ts := tsFoundry.Parse(idFoundry, line)
		HashH ^= ts.HashH()
		HashL ^= ts.HashL()
	}

	require.Equal(b, count, b.N)
}
