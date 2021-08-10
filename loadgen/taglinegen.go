package loadgen

import (
	"bufio"
	"bytes"
	"io"
	"math/rand"
	"os/exec"
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
		i := 0
		for l := range g.inner.GetLines() {
			i++
			if i >= g.n {
				break
			}
			g.lines = append(g.lines, l)
		}
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

// CmdTagLineGenerator runs a subcommand and provides each line it outputs as a
// tag line.  This allows offloading tagline generation to another process,
// allowing benchmarks to focus on allocation and performance only of the
// consumption functionality.
type CmdTagLineGenerator struct {
	name string
	n    int
}

// NewCmdTagLineGenerator creates a generator that will return `n` lines
// from the named command in `loadgen/cmd`.
func NewCmdTagLineGenerator(name string, n int) *CmdTagLineGenerator {
	return &CmdTagLineGenerator{name, n}
}

func (g *CmdTagLineGenerator) GetLines() chan ([]byte) {
	reader, writer := io.Pipe()
	gobin, err := exec.LookPath("go")
	cmd := exec.Cmd{
		Path: gobin,
		// TODO: assumes this is running from the root of this project
		Args:   []string{"go", "run", "./loadgen/cmd/" + g.name},
		Stdout: writer,
	}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(reader)

	// we reuse buffers after they are likely to have been consumed, by keeping
	// more buffers than there are slots in the channel
	const bufsize = 512
	const numbufs = 512
	const chansize = 500 // must be < numbufs
	bufs := make([][]byte, 0, numbufs)
	bufidx := 0

	for i := 0; i < numbufs; i++ {
		bufs = append(bufs, make([]byte, 0, bufsize))
	}
	c := make(chan ([]byte), chansize)

	ready := make(chan (struct{}))

	go func() {
		loop := func(n int) {
			i := 0
			for scanner.Scan() {
				if i >= n {
					break
				}
				line := scanner.Bytes()

				// make a local copy, since scanner will overwrite its buffer
				buf := bufs[bufidx%numbufs]
				bufidx++
				buf = buf[:len(line)]
				copy(buf, line)

				c <- buf

				i++
			}
		}

		prefill := chansize / 2
		if prefill > g.n {
			prefill = g.n
		}

		// fill the buffer halfway before signalling that we are ready; this allows
		// time for the process to start, and ensures we're not struggling to fill the
		// buffer for the first few lines.
		loop(prefill)

		// signal readiness
		close(ready)

		// finish up the g.n lines
		loop(g.n - prefill)

		// and signal the end of the lines
		close(c)

		// try to kill the command if it's not already killed
		_ = cmd.Process.Kill()
	}()

	// wait until the goroutine says it's ready before returning the channel
	<-ready

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
