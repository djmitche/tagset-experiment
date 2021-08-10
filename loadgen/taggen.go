package loadgen

import (
	"fmt"
	"math/rand"
)

// ---

// A TagGenerator provides a channel full of tags.  In general, `GetTags` can be
// profitably called multiple times for the same instance.
type TagGenerator interface {
	GetTags() chan ([]byte)
}

// ---

// A LowCardinalityTagGenerator generates tags from a given set
type LowCardinalityTagGenerator struct {
	choices [][]byte
}

func NewLowCardinalityTagGenerator(choices [][]byte) *LowCardinalityTagGenerator {
	return &LowCardinalityTagGenerator{choices}
}

func (tg *LowCardinalityTagGenerator) GetTags() chan ([]byte) {
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

// A HighCardinalityTagGenerator generates tags using a given suffix and random
// numeric suffixes of the given cardinality
type HighCardinalityTagGenerator struct {
	prefix      string
	cardinality int
}

func NewHighCardinalityTagGenerator(prefix string, cardinality int) *HighCardinalityTagGenerator {
	return &HighCardinalityTagGenerator{prefix, cardinality}
}

func (tg *HighCardinalityTagGenerator) GetTags() chan ([]byte) {
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

// A HostnameTagGenerator generates random hostname tags.  The hostnames look
// roughly like EC2 instances.
type HostnameTagGenerator struct{}

func NewHostnameTagGenerator() *HostnameTagGenerator {
	return &HostnameTagGenerator{}
}

func (tg *HostnameTagGenerator) GetTags() chan ([]byte) {
	c := make(chan ([]byte), 512)
	go func() {
		for {
			c <- []byte(fmt.Sprintf("host:i-%x", rand.Uint32()))
		}
	}()
	return c
}
