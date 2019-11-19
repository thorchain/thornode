package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type EventSuite struct{}

var _ = Suite(&EventSuite{})

func (s EventSuite) TestSwapEvent(c *C) {
	evt := NewEventSwap(
		common.BNBAsset,
		sdk.NewUint(5),
		sdk.NewUint(5),
		sdk.NewDec(5),
	)
	c.Check(evt.Type(), Equals, "swap")
}

func (s EventSuite) TestStakeEvent(c *C) {
	evt := NewEventStake(
		common.BNBAsset,
		sdk.NewUint(5),
	)
	c.Check(evt.Type(), Equals, "stake")
}

func (s EventSuite) TestUnstakeEvent(c *C) {
	evt := NewEventUnstake(
		common.BNBAsset,
		sdk.NewUint(6),
		5000,
		sdk.NewDec(0),
	)
	c.Check(evt.Type(), Equals, "unstake")
}

func (s EventSuite) TestPool(c *C) {
	evt := NewEventPool(common.BNBAsset, Enabled)
	c.Check(evt.Type(), Equals, "pool")
	c.Check(evt.Pool.String(), Equals, common.BNBAsset.String())
	c.Check(evt.Status.String(), Equals, Enabled.String())
}

func (s EventSuite) TestAdminConfig(c *C) {
	evt := NewEventAdminConfig("foo", "bar")
	c.Check(evt.Type(), Equals, "admin_config")
	c.Check(evt.Key, Equals, "foo")
	c.Check(evt.Value, Equals, "bar")
}

func (s EventSuite) TestEvent(c *C) {
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	swap := NewEventSwap(
		common.BNBAsset,
		sdk.NewUint(5),
		sdk.NewUint(5),
		sdk.NewDec(5),
	)

	swapBytes, _ := json.Marshal(swap)
	evt := NewEvent(
		swap.Type(),
		12,
		common.NewTx(
			txID,
			GetRandomBNBAddress(),
			GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(common.BNBAsset, sdk.NewUint(320000000)),
				common.NewCoin(common.RuneAsset(), sdk.NewUint(420000000)),
			},
			"SWAP:BNB.BNB",
		),
		swapBytes,
		Success,
	)

	c.Check(evt.Empty(), Equals, false)

	txID, err = common.NewTxID("B1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	stake := NewEventStake(
		common.BNBAsset,
		sdk.NewUint(5),
	)
	stakeBytes, _ := json.Marshal(stake)
	evt2 := NewEvent(
		stake.Type(),
		12,
		common.NewTx(
			txID,
			GetRandomBNBAddress(),
			GetRandomBNBAddress(),
			common.Coins{
				common.NewCoin(common.BNBAsset, sdk.NewUint(320000000)),
				common.NewCoin(common.RuneAsset(), sdk.NewUint(420000000)),
			},
			"SWAP:BNB.BNB",
		),
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
	emptyRefundEvent := NewEmptyRefundEvent()
	c.Check(emptyRefundEvent.Type(), Equals, "empty-refund")
}
