package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type StakerPoolSuite struct{}

var _ = Suite(&StakerPoolSuite{})

func (StakerPoolSuite) TestStakerPool(c *C) {
	bnb, err := common.NewAddress("bnb1xlvns0n2mxh77mzaspn2hgav4rr4m8eerfju38")
	c.Assert(err, IsNil)

	stakerPool := NewStakerPool(bnb)
	c.Check(stakerPool.RuneAddress, Equals, bnb)
	c.Assert(stakerPool.PoolUnits, NotNil)

	stakerPoolItem := &StakerPoolItem{
		Asset: common.BNBAsset,
		Units: sdk.NewUint(100),
	}
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	txID2, err := common.NewTxID("47B4FE474A63DDF79DF2790C1C5162F4C213484750AB8292CFE7342E4B0B40E2")
	c.Assert(err, IsNil)
	stakerPoolItem.AddStakerTxDetail(txID, sdk.NewUint(100), sdk.NewUint(100))
	stakerPool.UpsertStakerPoolItem(stakerPoolItem)
	c.Assert(stakerPool.PoolUnits, NotNil)
	c.Assert(len(stakerPool.PoolUnits), Equals, 1)
	stakerPoolItem.AddStakerTxDetail(txID2, sdk.NewUint(50), sdk.NewUint(50))
	c.Assert(len(stakerPool.PoolUnits), Equals, 1)
	c.Assert(stakerPoolItem.StakeDetails, NotNil)
	spi := stakerPool.GetStakerPoolItem(common.BNBAsset)
	c.Assert(spi, NotNil)
	c.Assert(len(spi.StakeDetails), Equals, 2)
	stakerPool.RemoveStakerPoolItem(common.RuneA1FAsset)
	stakerPool.RemoveStakerPoolItem(common.BNBAsset)
	c.Assert(len(stakerPool.PoolUnits), Equals, 0)
	stakerPool.UpsertStakerPoolItem(stakerPoolItem)
	stakerPoolItem1 := &StakerPoolItem{
		Asset: common.RuneB1AAsset,
		Units: sdk.NewUint(100),
	}
	stakerPool.UpsertStakerPoolItem(stakerPoolItem1)
	stakerPool.UpsertStakerPoolItem(stakerPoolItem)
	stakerPool.RemoveStakerPoolItem(common.RuneB1AAsset)
	c.Assert(len(stakerPool.PoolUnits), Equals, 1)
	c.Log(stakerPool.String())
	spi1 := stakerPool.GetStakerPoolItem(common.RuneA1FAsset)
	c.Assert(spi1.Units.IsZero(), Equals, true)
}
