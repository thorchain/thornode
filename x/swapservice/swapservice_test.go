package swapservice

import (
	"math"
	"testing"

	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type SwapServiceSuite struct{}

var _ = Suite(&SwapServiceSuite{})

func round(input float64) float64 {
	return math.Round(input*100) / 100
}
