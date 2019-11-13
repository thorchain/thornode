package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type VaultSuite struct{}

var _ = Suite(&VaultSuite{})

func (s VaultSuite) TestCalcBlockRewards(c *C) {
	bondR, poolR := calcBlockRewards(sdk.NewUint(1000 * common.One))
	c.Check(bondR.Uint64(), Equals, uint64(2114), Commentf("%d", bondR.Uint64()))
	c.Check(poolR.Uint64(), Equals, uint64(1056), Commentf("%d", poolR.Uint64()))
}

func (s VaultSuite) TestCalcPoolRewards(c *C) {
	p1 := NewPool()
	p1.BalanceRune = sdk.NewUint(100 * common.One)
	p2 := NewPool()
	p2.BalanceRune = sdk.NewUint(50 * common.One)
	pools := []Pool{p1, p2}

	totalRune := sdk.NewUint(150 * common.One)
	totalRewards := sdk.NewUint(60 * common.One)
	amts := calcPoolRewards(totalRewards, totalRune, pools)
	c.Assert(amts, HasLen, 2)
	c.Check(amts[0].Equal(sdk.NewUint(40*common.One)), Equals, true, Commentf("%d", amts[0].Uint64()))
	c.Check(amts[1].Equal(sdk.NewUint(20*common.One)), Equals, true, Commentf("%d", amts[1].Uint64()))
	c.Check(amts[0].Add(amts[1]).Equal(totalRewards), Equals, true)

	p1 = NewPool()
	p1.BalanceRune = sdk.NewUint(114.265 * common.One)
	p2 = NewPool()
	p2.BalanceRune = sdk.NewUint(23.875 * common.One)
	pools = []Pool{p1, p2}

	totalRune = sdk.NewUint(138.14 * common.One)
	totalRewards = sdk.NewUint(12.45 * common.One)
	amts = calcPoolRewards(totalRewards, totalRune, pools)
	c.Assert(amts, HasLen, 2)
	c.Check(amts[0].Equal(sdk.NewUint(1029824272)), Equals, true, Commentf("%d", amts[0].Uint64()))
	c.Check(amts[1].Equal(sdk.NewUint(215175728)), Equals, true, Commentf("%d", amts[1].Uint64()))
	c.Check(amts[0].Add(amts[1]).Equal(totalRewards), Equals, true, Commentf("%d", amts[0].Add(amts[1]).Uint64()))

}
