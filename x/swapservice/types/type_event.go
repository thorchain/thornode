package types

import (
	"encoding/json"

	common "gitlab.com/thorchain/bepswap/common"
)

type Event struct {
	ID      int64       `json:"id"`
	Type    string      `json:"type"`
	InHash  common.TxID `json:"in_hash"`
	OutHash common.TxID `json:"out_hash"`
	// Should we have timestamps and addresses if they are available via the
	// binance API?
	// InStamp    time.Time         `json:"in_timestamp"`
	// OutStamp   time.Time         `json:"out_timestamp"`
	// InAddress  common.BnbAddress `json:"in_address"`
	// OutAddress common.BnbAddress `json:"out_address"`
	Pool  common.Ticker   `json:"pool"`
	Event json.RawMessage `json:"event"`
}

func NewEvent(typ string, inHash common.TxID, pool common.Ticker, evt json.RawMessage) Event {
	return Event{
		Type:   typ,
		InHash: inHash,
		Pool:   pool,
		Event:  evt,
	}
}

func (evt Event) Empty() bool {
	return evt.InHash.IsEmpty()
}

type Events []Event

func (evts Events) GetByInHash(txID common.TxID) Event {
	for _, evt := range evts {
		if evt.InHash.Equals(txID) {
			return evt
		}
	}
	return Event{}
}

type EventSwap struct {
	SourceCoin common.Coin   `json:"source_coin"`
	TargetCoin common.Coin   `json:"target_coin"`
	Slip       common.Amount `json:"slip"`
}

func NewEventSwap(s, t common.Coin, slip common.Amount) EventSwap {
	return EventSwap{
		SourceCoin: s,
		TargetCoin: t,
		Slip:       slip,
	}
}

func (e EventSwap) Type() string {
	return "swap"
}

type EventStake struct {
	RuneAmount  common.Amount `json:"rune_amount"`
	TokenAmount common.Amount `json:"token_amount"`
	StakeUnits  common.Amount `json:"stake_units"`
}

func NewEventStake(r, t, s common.Amount) EventStake {
	return EventStake{
		RuneAmount:  r,
		TokenAmount: t,
		StakeUnits:  s,
	}
}

func (e EventStake) Type() string {
	return "stake"
}

type EventUnstake struct {
	RuneAmount  common.Amount `json:"rune_amount"`
	TokenAmount common.Amount `json:"token_amount"`
	StakeUnits  common.Amount `json:"stake_units"`
}

func NewEventUnstake(r, t, s common.Amount) EventUnstake {
	return EventUnstake{
		RuneAmount:  r,
		TokenAmount: t,
		StakeUnits:  s,
	}
}

func (e EventUnstake) Type() string {
	return "unstake"
}
