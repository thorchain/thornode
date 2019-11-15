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

	bondR, poolR = calcBlockRewards(sdk.ZeroUint())
	c.Check(bondR.Uint64(), Equals, uint64(0), Commentf("%d", bondR.Uint64()))
	c.Check(poolR.Uint64(), Equals, uint64(0), Commentf("%d", poolR.Uint64()))
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

	// Check that we don't error when there are no rewards
	p1 = NewPool()
	p1.BalanceRune = sdk.NewUint(114.265 * common.One)
	p2 = NewPool()
	p2.BalanceRune = sdk.NewUint(23.875 * common.One)
	pools = []Pool{p1, p2}

	totalRune = sdk.NewUint(138.14 * common.One)
	totalRewards = sdk.ZeroUint()
	amts = calcPoolRewards(totalRewards, totalRune, pools)
	c.Assert(amts, HasLen, 2)
	c.Check(amts[0].IsZero(), Equals, true, Commentf("%d", amts[0].Uint64()))
	c.Check(amts[1].IsZero(), Equals, true, Commentf("%d", amts[1].Uint64()))
}

func (s VaultSuite) TestCalcNodeRewards(c *C) {
	blocks := sdk.NewUint(5)
	totalUnits := sdk.NewUint(100)
	totalReward := sdk.NewUint(3000)
	reward := calcNodeRewards(blocks, totalUnits, totalReward)
	c.Check(reward.Uint64(), Equals, uint64(150))

	blocks = sdk.NewUint(78)
	totalUnits = sdk.NewUint(7357)
	totalReward = sdk.NewUint(275.357 * common.One)
	reward = calcNodeRewards(blocks, totalUnits, totalReward)
	c.Check(reward.Uint64(), Equals, uint64(291937556))

	// check for no rewards
	blocks = sdk.NewUint(78)
	totalUnits = sdk.NewUint(7357)
	totalReward = sdk.ZeroUint()
	reward = calcNodeRewards(blocks, totalUnits, totalReward)
	c.Check(reward.Uint64(), Equals, uint64(0))
}