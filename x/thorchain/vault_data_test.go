package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"

	. "gopkg.in/check.v1"
)

type VaultSuite struct{}

var _ = Suite(&VaultSuite{})

func (s VaultSuite) TestCalcBlockRewards(c *C) {
	ver := constants.SWVersion
	constAccessor := constants.GetConstantValues(ver)
	emissionCurve := constAccessor.GetInt64Value(constants.EmissionCurve)
	blocksPerYear := constAccessor.GetInt64Value(constants.BlocksPerYear)
	bondR, poolR, stakerD := calcBlockRewards(sdk.NewUint(1000*common.One), sdk.NewUint(2000*common.One), sdk.NewUint(1000*common.One), sdk.ZeroUint(), emissionCurve, blocksPerYear)
	c.Check(bondR.Uint64(), Equals, uint64(1761), Commentf("%d", bondR.Uint64()))
	c.Check(poolR.Uint64(), Equals, uint64(880), Commentf("%d", poolR.Uint64()))
	c.Check(stakerD.Uint64(), Equals, uint64(0), Commentf("%d", poolR.Uint64()))

	bondR, poolR, stakerD = calcBlockRewards(sdk.NewUint(1000*common.One), sdk.NewUint(2000*common.One), sdk.NewUint(1000*common.One), sdk.NewUint(3000), emissionCurve, blocksPerYear)
	c.Check(bondR.Uint64(), Equals, uint64(3761), Commentf("%d", bondR.Uint64()))
	c.Check(poolR.Uint64(), Equals, uint64(0), Commentf("%d", poolR.Uint64()))
	c.Check(stakerD.Uint64(), Equals, uint64(1120), Commentf("%d", poolR.Uint64()))

	bondR, poolR, stakerD = calcBlockRewards(sdk.NewUint(1000*common.One), sdk.NewUint(2000*common.One), sdk.ZeroUint(), sdk.ZeroUint(), emissionCurve, blocksPerYear)
	c.Check(bondR.Uint64(), Equals, uint64(0), Commentf("%d", bondR.Uint64()))
	c.Check(poolR.Uint64(), Equals, uint64(0), Commentf("%d", poolR.Uint64()))
	c.Check(stakerD.Uint64(), Equals, uint64(0), Commentf("%d", poolR.Uint64()))

	bondR, poolR, stakerD = calcBlockRewards(sdk.NewUint(1000*common.One), sdk.NewUint(1000*common.One), sdk.NewUint(1000*common.One), sdk.ZeroUint(), emissionCurve, blocksPerYear)
	c.Check(bondR.Uint64(), Equals, uint64(2641), Commentf("%d", bondR.Uint64()))
	c.Check(poolR.Uint64(), Equals, uint64(0), Commentf("%d", poolR.Uint64()))
	c.Check(stakerD.Uint64(), Equals, uint64(0), Commentf("%d", poolR.Uint64()))

	bondR, poolR, stakerD = calcBlockRewards(sdk.ZeroUint(), sdk.NewUint(1000*common.One), sdk.NewUint(1000*common.One), sdk.ZeroUint(), emissionCurve, blocksPerYear)
	c.Check(bondR.Uint64(), Equals, uint64(0), Commentf("%d", bondR.Uint64()))
	c.Check(poolR.Uint64(), Equals, uint64(2641), Commentf("%d", poolR.Uint64()))
	c.Check(stakerD.Uint64(), Equals, uint64(0), Commentf("%d", poolR.Uint64()))

	bondR, poolR, stakerD = calcBlockRewards(sdk.NewUint(2001*common.One), sdk.NewUint(1000*common.One), sdk.NewUint(1000*common.One), sdk.ZeroUint(), emissionCurve, blocksPerYear)
	c.Check(bondR.Uint64(), Equals, uint64(2641), Commentf("%d", bondR.Uint64()))
	c.Check(poolR.Uint64(), Equals, uint64(0), Commentf("%d", poolR.Uint64()))
	c.Check(stakerD.Uint64(), Equals, uint64(0), Commentf("%d", poolR.Uint64()))
}

func (s VaultSuite) TestCalcPoolDeficit(c *C) {
	pool1Fees := sdk.NewUint(1000)
	pool2Fees := sdk.NewUint(3000)
	totalFees := sdk.NewUint(4000)

	stakerDeficit := sdk.NewUint(1120)
	amt1 := calcPoolDeficit(stakerDeficit, totalFees, pool1Fees)
	amt2 := calcPoolDeficit(stakerDeficit, totalFees, pool2Fees)

	c.Check(amt1.Equal(sdk.NewUint(280)), Equals, true, Commentf("%d", amt1.Uint64()))
	c.Check(amt2.Equal(sdk.NewUint(840)), Equals, true, Commentf("%d", amt2.Uint64()))
}
