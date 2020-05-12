package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type EventSuite struct{}

var _ = Suite(&EventSuite{})

func (s EventSuite) TestSwapEvent(c *C) {
	evt := NewEventSwap(
		common.BNBAsset,
		sdk.NewUint(5),
		sdk.NewUint(5),
		sdk.NewUint(5),
		sdk.ZeroUint(),
		GetRandomTx(),
	)
	c.Check(evt.Type(), Equals, "swap")
}

func (s EventSuite) TestStakeEvent(c *C) {
	evt := NewEventStake(
		common.BNBAsset,
		sdk.NewUint(5),
		GetRandomTx(),
	)
	c.Check(evt.Type(), Equals, "stake")
}

func (s EventSuite) TestUnstakeEvent(c *C) {
	evt := NewEventUnstake(
		common.BNBAsset,
		sdk.NewUint(6),
		5000,
		sdk.NewDec(0),
		GetRandomTx(),
	)
	c.Check(evt.Type(), Equals, "unstake")
}

func (s EventSuite) TestPool(c *C) {
	evt := NewEventPool(common.BNBAsset, Enabled)
	c.Check(evt.Type(), Equals, "pool")
	c.Check(evt.Pool.String(), Equals, common.BNBAsset.String())
	c.Check(evt.Status.String(), Equals, Enabled.String())
}

func (s EventSuite) TestReward(c *C) {
	evt := NewEventRewards(sdk.NewUint(300), []PoolAmt{
		{common.BNBAsset, 30},
		{common.BTCAsset, 40},
	})
	c.Check(evt.Type(), Equals, "rewards")
	c.Check(evt.BondReward.String(), Equals, "300")
	c.Assert(evt.PoolRewards, HasLen, 2)
	c.Check(evt.PoolRewards[0].Asset.Equals(common.BNBAsset), Equals, true)
	c.Check(evt.PoolRewards[0].Amount, Equals, int64(30))
	c.Check(evt.PoolRewards[1].Asset.Equals(common.BTCAsset), Equals, true)
	c.Check(evt.PoolRewards[1].Amount, Equals, int64(40))
}

func (s EventSuite) TestEvent(c *C) {
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	txIn := common.NewTx(
		txID,
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(320000000)),
			common.NewCoin(common.RuneAsset(), sdk.NewUint(420000000)),
		},
		BNBGasFeeSingleton,
		"SWAP:BNB.BNB",
	)
	swap := NewEventSwap(
		common.BNBAsset,
		sdk.NewUint(5),
		sdk.NewUint(5),
		sdk.NewUint(5),
		sdk.ZeroUint(),
		txIn,
	)

	swapBytes, _ := json.Marshal(swap)
	evt := NewEvent(swap.Type(),
		12,
		txIn,
		swapBytes,
		Success,
	)

	c.Check(evt.Empty(), Equals, false)

	txID, err = common.NewTxID("B1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	txIn = common.NewTx(
		txID,
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(320000000)),
			common.NewCoin(common.RuneAsset(), sdk.NewUint(420000000)),
		},
		BNBGasFeeSingleton,
		"SWAP:BNB.BNB",
	)
	stake := NewEventStake(
		common.BNBAsset,
		sdk.NewUint(5),
		txIn,
	)
	stakeBytes, _ := json.Marshal(stake)
	evt2 := NewEvent(stake.Type(),
		12,
		txIn,
		stakeBytes,
		Success,
	)

	events := Events{evt, evt2}
	found, events := events.PopByInHash(txID)
	c.Assert(found, HasLen, 1)
	c.Check(found[0].Empty(), Equals, false)
	c.Check(found[0].Type, Equals, evt2.Type)
	c.Assert(events, HasLen, 1)
	c.Check(events[0].Type, Equals, evt.Type)

	c.Check(Event{}.Empty(), Equals, true)
}

func (s EventSuite) TestSlash(c *C) {
	evt := NewEventSlash(common.BNBAsset, []PoolAmt{
		{common.BNBAsset, -20},
		{common.RuneAsset(), 30},
	})
	c.Check(evt.Type(), Equals, "slash")
	c.Check(evt.Pool, Equals, common.BNBAsset)
	c.Assert(evt.SlashAmount, HasLen, 2)
	c.Check(evt.SlashAmount[0].Asset, Equals, common.BNBAsset)
	c.Check(evt.SlashAmount[0].Amount, Equals, int64(-20))
	c.Check(evt.SlashAmount[1].Asset, Equals, common.RuneAsset())
	c.Check(evt.SlashAmount[1].Amount, Equals, int64(30))
}

func (s EventSuite) TestEventGas(c *C) {
	eg := NewEventGas()
	c.Assert(eg, NotNil)
	eg.UpsertGasPool(GasPool{
		Asset:    common.BNBAsset,
		AssetAmt: sdk.NewUint(1000),
		RuneAmt:  sdk.ZeroUint(),
	})
	c.Assert(eg.Pools, HasLen, 1)
	c.Assert(eg.Pools[0].Asset, Equals, common.BNBAsset)
	c.Assert(eg.Pools[0].RuneAmt.Equal(sdk.ZeroUint()), Equals, true)
	c.Assert(eg.Pools[0].AssetAmt.Equal(sdk.NewUint(1000)), Equals, true)

	eg.UpsertGasPool(GasPool{
		Asset:    common.BNBAsset,
		AssetAmt: sdk.NewUint(1234),
		RuneAmt:  sdk.NewUint(1024),
	})
	c.Assert(eg.Pools, HasLen, 1)
	c.Assert(eg.Pools[0].Asset, Equals, common.BNBAsset)
	c.Assert(eg.Pools[0].RuneAmt.Equal(sdk.NewUint(1024)), Equals, true)
	c.Assert(eg.Pools[0].AssetAmt.Equal(sdk.NewUint(2234)), Equals, true)

	eg.UpsertGasPool(GasPool{
		Asset:    common.BTCAsset,
		AssetAmt: sdk.NewUint(1024),
		RuneAmt:  sdk.ZeroUint(),
	})
	c.Assert(eg.Pools, HasLen, 2)
	c.Assert(eg.Pools[1].Asset, Equals, common.BTCAsset)
	c.Assert(eg.Pools[1].AssetAmt.Equal(sdk.NewUint(1024)), Equals, true)
	c.Assert(eg.Pools[1].RuneAmt.Equal(sdk.ZeroUint()), Equals, true)

	eg.UpsertGasPool(GasPool{
		Asset:    common.BTCAsset,
		AssetAmt: sdk.ZeroUint(),
		RuneAmt:  sdk.ZeroUint(),
	})

	c.Assert(eg.Pools, HasLen, 2)
	c.Assert(eg.Pools[1].Asset, Equals, common.BTCAsset)
	c.Assert(eg.Pools[1].AssetAmt.Equal(sdk.NewUint(1024)), Equals, true)
	c.Assert(eg.Pools[1].RuneAmt.Equal(sdk.ZeroUint()), Equals, true)

	eg.UpsertGasPool(GasPool{
		Asset:    common.BTCAsset,
		AssetAmt: sdk.ZeroUint(),
		RuneAmt:  sdk.NewUint(3333),
	})

	c.Assert(eg.Pools, HasLen, 2)
	c.Assert(eg.Pools[1].Asset, Equals, common.BTCAsset)
	c.Assert(eg.Pools[1].AssetAmt.Equal(sdk.NewUint(1024)), Equals, true)
	c.Assert(eg.Pools[1].RuneAmt.Equal(sdk.NewUint(3333)), Equals, true)
}

func (s EventSuite) TestEventFee(c *C) {
	event := NewEventFee(GetRandomTxHash(), common.Fee{
		Coins: common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(1024)),
		},
		PoolDeduct: sdk.NewUint(1023),
	})
	c.Assert(event.Type(), Equals, FeeEventType)
	evts, err := event.Events()
	c.Assert(err, IsNil)
	c.Assert(evts, HasLen, 1)
}
