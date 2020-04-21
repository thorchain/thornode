package common

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type CommonSuite struct{}

var _ = Suite(&CommonSuite{})

func (s CommonSuite) TestGetShare(c *C) {
	part := sdk.NewUint(149506590)
	total := sdk.NewUint(50165561086)
	alloc := sdk.NewUint(50000000)
	share := GetShare(part, total, alloc)
	c.Assert(share.Equal(sdk.NewUint(149013)), Equals, true)
}
