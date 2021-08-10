package loadgen

import (
	"bytes"
	"log"
	"math/rand"
)

// ---

// A TagLineGenerator provides a channel full of comma-separated tag lines.  In
// general a TagLineGenerator's `GetLines` method can only be called once, with
// exceptions as noted.
type TagLineGenerator interface {
	// Get a channel that will provide lines full of comma-seprated tags
	GetLines() chan ([]byte)
}

// ---

// PreflightTagLineGenerator buffers all the lines from the inner generator.
// Use this for benchmarks.  This generator can be used repeatedly, such as in
// different benchmarks.
type PreflightTagLineGenerator struct {
	inner TagLineGenerator
	n     int
	lines [][]byte
}

// NewPreflightTagLineGenerator creates a generator that will return `n` lines
// derived from the given tagLineGenerator.
func NewPreflightTagLineGenerator(n int, inner TagLineGenerator) *PreflightTagLineGenerator {
	return &PreflightTagLineGenerator{
		inner: inner,
		n:     n,
		lines: [][]byte{},
	}
}

// Preflight generates and caches the output from the wrapped generator.
func (g *PreflightTagLineGenerator) Preflight() {
	if len(g.lines) == 0 {
		log.Printf("Pre-flighting %d tag lines", g.n)
		i := 0
		for l := range g.inner.GetLines() {
			i++
			if i >= g.n {
				break
			}
			g.lines = append(g.lines, l)
		}
		log.Printf("preflight complete")
	}
}

func (g *PreflightTagLineGenerator) GetLines() chan ([]byte) {
	g.Preflight()

	// now push the cached lines as quickly as possible
	c := make(chan ([]byte), 512)
	go func() {
		for _, l := range g.lines {
			c <- l
		}
		close(c)
	}()

	return c
}

// ---

// A DuplicateTagLineGenerator wraps another generator and duplicates lines
// from that generator randomly.  This is intended to simulate many values
// of the same context being sent.
type DuplicateTagLineGenerator struct {
	// underlying source of tag lines
	inner TagLineGenerator

	// number of distinct duplicates to handle at any time
	numDistinct int

	// maximum number of times to duplicate a line
	maxTimes int
}

func NewDuplicateTagLineGenerator(inner TagLineGenerator, numDistinct, maxTimes int) *DuplicateTagLineGenerator {
	return &DuplicateTagLineGenerator{inner, numDistinct, maxTimes}
}

func (g *DuplicateTagLineGenerator) GetLines() chan ([]byte) {
	// the upstream channel
	upstream := g.inner.GetLines()

	// a cache of log lines to duplicate (len == numDistinct)
	cache := make([][]byte, g.numDistinct)

	// duplicates remaining for each of those
	timesRemaining := make([]int, g.numDistinct)

	fill := func(i int) {
		cache[i] = <-upstream
		timesRemaining[i] = rand.Intn(g.maxTimes-1) + 1
	}

	// pre-fill the cache..
	for i := range cache {
		fill(i)
	}

	c := make(chan ([]byte), 512)
	go func() {
		for {
			i := rand.Intn(len(cache))
			c <- cache[i]
			timesRemaining[i]--
			if timesRemaining[i] <= 0 {
				fill(i)
			}
		}
	}()

	return c
}

// ---

// A MultiplexingTagLineGenerator wraps other generators and mixes their output
// together fairly but randomly.
type MultiplexingTagLineGenerator struct {
	// underlying sources of tag lines
	inners []TagLineGenerator
}

func NewMultiplexingTagLineGenerator(inners []TagLineGenerator) *MultiplexingTagLineGenerator {
	return &MultiplexingTagLineGenerator{inners}
}

func (g *MultiplexingTagLineGenerator) GetLines() chan ([]byte) {
	sources := []chan ([]byte){}
	for _, inner := range g.inners {
		sources = append(sources, inner.GetLines())
	}

	c := make(chan ([]byte), 512)
	go func() {
		for {
			i := rand.Intn(len(sources))
			c <- <-sources[i]
		}
	}()

	return c
}

// ---

// A RandomTagLineGenerator generates random tag lines based on a set of tag generators.
type RandomTagLineGenerator struct {
	tagGenerators []RandomTagLineOptions
}

type RandomTagLineOptions struct {
	// generator generating the tags
	tg TagGenerator
	// probability of this generator being used (each repeat, as percent)
	prob int
	// max number of repeats of this generator per line
	repeats int
}

func NewRandomTagLineGenerator(tagGenerators []RandomTagLineOptions) *RandomTagLineGenerator {
	return &RandomTagLineGenerator{
		tagGenerators,
	}
}

func (g *RandomTagLineGenerator) GetLines() chan ([]byte) {
	var commaSep = []byte(",")
	c := make(chan ([]byte), 512)

	go func() {
		tagChans := make([]chan ([]byte), 0, len(g.tagGenerators))
		for _, tlo := range g.tagGenerators {
			c := tlo.tg.GetTags()
			tagChans = append(tagChans, c)
		}
		for {
			tags := make([][]byte, 0, len(tagChans))
			for i, ch := range tagChans {
				tlo := g.tagGenerators[i]
				for j := 0; j < tlo.repeats; j++ {
					if rand.Intn(100) < tlo.prob {
						tags = append(tags, <-ch)
					}
				}
			}
			c <- bytes.Join(tags, commaSep)
		}
	}()

	return c
}
