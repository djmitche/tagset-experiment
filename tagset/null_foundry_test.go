package tagset

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type NullFoundrySuite struct {
	FoundrySuite
}

func TestNullFoundry(t *testing.T) {
	suite.Run(t, &NullFoundrySuite{
		FoundrySuite: FoundrySuite{
			f: NewNullFoundry(),
		},
	})
}
