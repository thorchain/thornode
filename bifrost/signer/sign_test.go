package signer

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type SignSuite struct{}

var _ = Suite(&SignSuite{})
