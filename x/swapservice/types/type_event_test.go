package types

import (
	"encoding/json"

	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type EventSuite struct{}

var _ = Suite(&EventSuite{})

func (s EventSuite) TestSwapEvent(c *C) {
	evt := NewEventSwap(
		common.NewCoin(common.Ticker("BNB"), common.Amount("3.2")),
		common.NewCoin(common.Ticker("RUNE"), common.Amount("4.2")),
		common.Amount("5"),
		common.Amount("5"),
		common.Amount("5"),
		common.Amount("5"),
		common.Amount("5"),
	)
	c.Check(evt.Type(), Equals, "swap")
}

func (s EventSuite) TestStakeEvent(c *C) {
	evt := NewEventStake(
		common.Amount("6"),
		common.Amount("7"),
		common.Amount("5"),
	)
	c.Check(evt.Type(), Equals, "stake")
}

func (s EventSuite) TestUnstakeEvent(c *C) {
	evt := NewEventUnstake(
		common.Amount("6"),
		common.Amount("7"),
		common.Amount("5"),
	)
	c.Check(evt.Type(), Equals, "unstake")
}

func (s EventSuite) TestEvent(c *C) {
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	swap := NewEventSwap(
		common.NewCoin(common.Ticker("BNB"), common.Amount("3.2")),
		common.NewCoin(common.Ticker("RUNE"), common.Amount("4.2")),
		common.Amount("6"),
		common.Amount("5"),
		common.Amount("5"),
		common.Amount("4"),
		common.Amount("3"),
	)
	swapBytes, _ := json.Marshal(swap)
	evt := NewEvent(
		swap.Type(),
		txID,
		common.Ticker("BNB"),
		swapBytes,
		Success,
	)

	c.Check(evt.Empty(), Equals, false)

	txID, err = common.NewTxID("B1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	stake := NewEventStake(
		common.Amount("6"),
		common.Amount("7"),
		common.Amount("5"),
	)
	stakeBytes, _ := json.Marshal(stake)
	evt2 := NewEvent(
		stake.Type(),
		txID,
		common.Ticker("BNB"),
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
