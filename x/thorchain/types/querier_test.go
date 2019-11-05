package types

import (
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type QueryResPoolsSuite struct{}

var _ = Suite(&QueryResPoolsSuite{})

func (QueryResPoolsSuite) TestQueryResPools(c *C) {
	p := NewPool()
	p.Asset = common.BNBAsset
	var qrp QueryResPools
	qrp = append(qrp, p)
	c.Check(qrp.String(), Equals, "BNB.BNB")

}
