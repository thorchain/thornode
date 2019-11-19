package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type RservesSuite struct{}

var _ = Suite(&RservesSuite{})

func (s *RservesSuite) TestReserveContributors(c *C) {
	addr := GetRandomBNBAddress()
	res := NewReserveContributor(
		addr,
		sdk.NewUint(32*common.One),
	)
	c.Check(res.Address.Equals(addr), Equals, true)
	c.Check(res.Amount.Equal(sdk.NewUint(32*common.One)), Equals, true)

	reses := ReserveContributors{res}

	res = NewReserveContributor(
		GetRandomBNBAddress(),
		sdk.NewUint(10*common.One),
	)

	reses = reses.Add(res)
	c.Assert(reses, HasLen, 2)
	c.Check(reses[1].Amount.Equal(sdk.NewUint(10*common.One)), Equals, true)
	reses = reses.Add(res)
	c.Assert(reses, HasLen, 2)
	c.Check(reses[1].Amount.Equal(sdk.NewUint(20*common.One)), Equals, true)
}
