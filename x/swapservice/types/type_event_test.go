package types

import (
	"encoding/json"

	common "gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type EventSuite struct{}

var _ = Suite(&EventSuite{})

func (s EventSuite) TestSwapEvent(c *C) {
	evt := NewEventSwap(
		common.NewCoin(common.Ticker("BNB"), common.Amount("3.2")),
		common.NewCoin(common.Ticker("RUNE"), common.Amount("4.2")),
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
		common.Amount("5"),
	)
	swapBytes, _ := json.Marshal(swap)
	evt := NewEvent(
		swap.Type(),
		txID,
		common.Ticker("BNB"),
		swapBytes,
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
	)

	events := Events{evt, evt2}
	event := events.GetByInHash(txID)
	c.Check(event.Empty(), Equals, false)
	c.Check(event.Type, Equals, evt2.Type)

	c.Check(Event{}.Empty(), Equals, true)
}
