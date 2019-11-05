package thorchain

import (
	"math"
	"testing"

	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type ThorchainSuite struct{}

var _ = Suite(&ThorchainSuite{})

func round(input float64) float64 {
	return math.Round(input*100) / 100
}
