package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

type Event struct {
	ID      common.Amount `json:"id"`
	Type    string        `json:"type"`
	InHash  common.TxID   `json:"in_hash"`
	OutHash common.TxID   `json:"out_hash"`
	// Should we have timestamps and addresses if they are available via the
	// binance API?
	// InStamp    time.Time         `json:"in_timestamp"`
	// OutStamp   time.Time         `json:"out_timestamp"`
	// InAddress  common.Address `json:"in_address"`
	// OutAddress common.Address `json:"out_address"`
	Pool   common.Asset    `json:"pool"`
	Event  json.RawMessage `json:"event"`
	Status EventStatus     `json:"status"`
}

func NewEvent(typ string, inHash common.TxID, pool common.Asset, evt json.RawMessage, status EventStatus) Event {
	return Event{
		Type:   typ,
		InHash: inHash,
		Pool:   pool,
		Event:  evt,
		Status: status,
	}
}

func (evt Event) Empty() bool {
	return evt.InHash.IsEmpty()
}

type Events []Event

// Pops an event out of the event list by hash ID
func (evts Events) PopByInHash(txID common.TxID) (found Events, events Events) {
	for _, evt := range evts {
		if evt.InHash.Equals(txID) {
			found = append(found, evt)
		} else {
			events = append(events, evt)
		}
	}
	return
}

type EventSwap struct {
	SourceCoin common.Coin `json:"source_coin"`
	TargetCoin common.Coin `json:"target_coin"`
	PriceSlip  sdk.Uint    `json:"price_slip"`
	TradeSlip  sdk.Uint    `json:"trade_slip"`
	PoolSlip   sdk.Uint    `json:"pool_slip"`
	OutputSlip sdk.Uint    `json:"output_slip"`
	Fee        sdk.Uint    `json:"fee"`
}

func NewEventSwap(s, t common.Coin, priceSlip, tradeSlip, poolSlip, outputSlip, fee sdk.Uint) EventSwap {
	return EventSwap{
		SourceCoin: s,
		TargetCoin: t,
		PriceSlip:  priceSlip,
		TradeSlip:  tradeSlip,
		PoolSlip:   poolSlip,
		OutputSlip: outputSlip,
		Fee:        fee,
	}
}

func (e EventSwap) Type() string {
	return "swap"
}

type EventStake struct {
	RuneAmount  sdk.Uint `json:"rune_amount"`
	AssetAmount sdk.Uint `json:"asset_amount"`
	StakeUnits  sdk.Uint `json:"stake_units"`
}

func NewEventStake(r, t, s sdk.Uint) EventStake {
	return EventStake{
		RuneAmount:  r,
		AssetAmount: t,
		StakeUnits:  s,
	}
}

func (e EventStake) Type() string {
	return "stake"
}

type EventUnstake struct {
	RuneAmount  sdk.Int `json:"rune_amount"`
	AssetAmount sdk.Int `json:"asset_amount"`
	StakeUnits  sdk.Int `json:"stake_units"`
}

func NewEventUnstake(r, t, s sdk.Uint) EventUnstake {
	return EventUnstake{
		RuneAmount:  sdk.NewInt(-1).Mul(sdk.NewInt(int64(r.Uint64()))),
		AssetAmount: sdk.NewInt(-1).Mul(sdk.NewInt(int64(t.Uint64()))),
		StakeUnits:  sdk.NewInt(-1).Mul(sdk.NewInt(int64(s.Uint64()))),
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
