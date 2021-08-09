package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
)

// ---

// A tagLineGenerator provides a channel full of comma-separated tag lines.  In
// general a tagLineGenerator's `getLines` method can only be called once, with
// exceptions as noted.
type tagLineGenerator interface {
	// Get a channel that will provide lines full of comma-seprated tags
	getLines() chan ([]byte)
}

// ---

// preflightTagLineGenerator buffers all the lines from the inner generator.
// Use this for benchmarks.  This generator can be used repeatedly, such as in
// different benchmarks.
type preflightTagLineGenerator struct {
	inner tagLineGenerator
	n     int
	lines [][]byte
}

// newPreflightTagLineGenerator creates a generator that will return `n` lines
// derived from the given tagLineGenerator.
func newPreflightTagLineGenerator(n int, inner tagLineGenerator) *preflightTagLineGenerator {
	return &preflightTagLineGenerator{
		inner: inner,
		n:     n,
		lines: [][]byte{},
	}
}

func (g *preflightTagLineGenerator) preflight() {
	if len(g.lines) == 0 {
		log.Printf("Pre-flighting %d tag lines", g.n)
		i := 0
		for l := range g.inner.getLines() {
			i++
			if i >= g.n {
				break
			}
			g.lines = append(g.lines, l)
		}
		log.Printf("preflight complete")
	}
}

func (g *preflightTagLineGenerator) getLines() chan ([]byte) {
	g.preflight()

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

// A duplicateTagLineGenerator wraps another generator and duplicates lines
// from that generator randomly.  This is intended to simulate many values
// of the same context being sent.
type duplicateTagLineGenerator struct {
	// underlying source of tag lines
	inner tagLineGenerator

	// number of distinct duplicates to handle at any time
	numDistinct int

	// maximum number of times to duplicate a line
	maxTimes int
}

func newDuplicateTagLineGenerator(inner tagLineGenerator, numDistinct, maxTimes int) *duplicateTagLineGenerator {
	return &duplicateTagLineGenerator{inner, numDistinct, maxTimes}
}

func (g *duplicateTagLineGenerator) getLines() chan ([]byte) {
	// the upstream channel
	upstream := g.inner.getLines()

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

// A multiplexingTagLineGenerator wraps other generators and mixes their output
// together fairly but randomly.
type multiplexingTagLineGenerator struct {
	// underlying sources of tag lines
	inners []tagLineGenerator
}

func newMultiplexingTagLineGenerator(inners []tagLineGenerator) *multiplexingTagLineGenerator {
	return &multiplexingTagLineGenerator{inners}
}

func (g *multiplexingTagLineGenerator) getLines() chan ([]byte) {
	sources := []chan ([]byte){}
	for _, inner := range g.inners {
		sources = append(sources, inner.getLines())
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

// A randomTagLineGenerator generates random tag lines based on a set of tag generators.
type randomTagLineGenerator struct {
	tagGenerators []randomTagLineOptions
}

type randomTagLineOptions struct {
	// generator generating the tags
	tg tagGenerator
	// probability of this generator being used (each repeat, as percent)
	prob int
	// max number of repeats of this generator per line
	repeats int
}

func newRandomTagLineGenerator(tagGenerators []randomTagLineOptions) *randomTagLineGenerator {
	return &randomTagLineGenerator{
		tagGenerators,
	}
}

func (g *randomTagLineGenerator) getLines() chan ([]byte) {
	var commaSep = []byte(",")
	c := make(chan ([]byte), 512)

	go func() {
		tagChans := make([]chan ([]byte), 0, len(g.tagGenerators))
		for _, tlo := range g.tagGenerators {
			c := tlo.tg.getTags()
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

// ---

// A tagGenerator provides a channel full of tags.  In general, `getTags` can be
// profitably called multiple times for the same instance.
type tagGenerator interface {
	getTags() chan ([]byte)
}

// ---

// A lowCardinalityTagGenerator generates tags from a given set
type lowCardinalityTagGenerator struct {
	choices [][]byte
}

func newLowCardinalityTagGenerator(choices [][]byte) *lowCardinalityTagGenerator {
	return &lowCardinalityTagGenerator{choices}
}

func (tg *lowCardinalityTagGenerator) getTags() chan ([]byte) {
	c := make(chan ([]byte), 512)
	go func() {
		for {
			i := rand.Intn(len(tg.choices))
			c <- append([]byte{}, tg.choices[i]...)
		}
	}()
	return c
}

// ---

// A highCardinalityTagGenerator generates tags using a given suffix and random
// numeric suffixes of the given cardinality
type highCardinalityTagGenerator struct {
	prefix      string
	cardinality int
}

func newHighCardinalityTagGenerator(prefix string, cardinality int) *highCardinalityTagGenerator {
	return &highCardinalityTagGenerator{prefix, cardinality}
}

func (tg *highCardinalityTagGenerator) getTags() chan ([]byte) {
	c := make(chan ([]byte), 512)
	go func() {
		for {
			i := rand.Intn(tg.cardinality)
			c <- []byte(fmt.Sprintf("%s:%x", tg.prefix, i))
		}
	}()
	return c
}

// ---

type hostnameTagGenerator struct{}

func newHostnameTagGenerator() *hostnameTagGenerator {
	return &hostnameTagGenerator{}
}

func (tg *hostnameTagGenerator) getTags() chan ([]byte) {
	c := make(chan ([]byte), 512)
	go func() {
		for {
			c <- []byte(fmt.Sprintf("host:i-%x", rand.Uint32()))
		}
	}()
	return c
}
