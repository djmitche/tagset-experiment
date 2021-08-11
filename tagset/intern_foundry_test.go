package tagset

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type InternFoundrySuite struct {
	FoundrySuite
}

func TestInternFoundry(t *testing.T) {
	suite.Run(t, &InternFoundrySuite{
		FoundrySuite: FoundrySuite{
			f: NewInternFoundry(),
		},
	})
}
