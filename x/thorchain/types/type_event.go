package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

type Event struct {
	ID     int64           `json:"id"`
	Height int64           `json:"height"`
	Type   string          `json:"type"`
	InTx   common.Tx       `json:"in_tx"`
	OutTx  common.Txs      `json:"out_txs"`
	Gas    common.Coins    `json:"gas"`
	Event  json.RawMessage `json:"event"`
	Status EventStatus     `json:"status"`
}

func NewEvent(typ string, ht int64, inTx common.Tx, evt json.RawMessage, status EventStatus) Event {
	return Event{
		Height: ht,
		Type:   typ,
		InTx:   inTx,
		Event:  evt,
		Status: status,
	}
}

func (evt Event) Empty() bool {
	return evt.InTx.ID.IsEmpty()
}

type Events []Event

// Pops an event out of the event list by hash ID
func (evts Events) PopByInHash(txID common.TxID) (found Events, events Events) {
	for _, evt := range evts {
		if evt.InTx.ID.Equals(txID) {
			found = append(found, evt)
		} else {
			events = append(events, evt)
		}
	}
	return
}

type EventSwap struct {
	Pool         common.Asset `json:"pool"`
	PriceTarget  sdk.Uint     `json:"price_target"`
	TradeSlip    sdk.Dec      `json:"trade_slip"`
	LiquidityFee sdk.Uint     `json:"liquidity_fee"`
}

func NewEventSwap(pool common.Asset, priceTarget, fee sdk.Uint, tradeSlip sdk.Dec) EventSwap {
	return EventSwap{
		Pool:         pool,
		PriceTarget:  priceTarget,
		TradeSlip:    tradeSlip,
		LiquidityFee: fee,
	}
}

func (e EventSwap) Type() string {
	return "swap"
}

type EventStake struct {
	Pool       common.Asset `json:"pool"`
	StakeUnits sdk.Uint     `json:"stake_units"`
}

func NewEventStake(pool common.Asset, su sdk.Uint) EventStake {
	return EventStake{
		Pool:       pool,
		StakeUnits: su,
	}
}

func (e EventStake) Type() string {
	return "stake"
}

type EventUnstake struct {
	Pool        common.Asset `json:"pool"`
	StakeUnits  sdk.Uint     `json:"stake_units"`
	BasisPoints int64        `json:"basis_points"` // 1 ==> 10,0000
	Asymmetry   sdk.Dec      `json:"asymmetry"`    // -1.0 <==> 1.0
}

func NewEventUnstake(pool common.Asset, su sdk.Uint, basisPts int64, asym sdk.Dec) EventUnstake {
	return EventUnstake{
		Pool:        pool,
		StakeUnits:  su,
		BasisPoints: basisPts,
		Asymmetry:   asym,
	}
}

func (e EventUnstake) Type() string {
	return "unstake"
}

type EmptyRefundEvent struct {
}

func NewEmptyRefundEvent() EmptyRefundEvent {
	return EmptyRefundEvent{}
}

func (e EmptyRefundEvent) Type() string {
	return "empty-refund"
}
