package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// Event bt
type Event struct {
	ID     int64           `json:"id"`
	Height int64           `json:"height"`
	Type   string          `json:"type"`
	InTx   common.Tx       `json:"in_tx"`
	OutTxs common.Txs      `json:"out_txs"`
	Gas    common.Coins    `json:"gas"`
	Event  json.RawMessage `json:"event"`
	Status EventStatus     `json:"status"`
}

const (
	SwapEventType        = `swap`
	StakeEventType       = `stake`
	UnstakeEventType     = `unstake`
	AdminConfigEventType = `admin_config`
	AddEventType         = `add`
	PoolEventType        = `pool`
	RewardEventType      = `rewards`
	EmptyRefundEventType = `empty-refund`
)

// NewEvent create a new  event
func NewEvent(id int64, typ string, ht int64, inTx common.Tx, evt json.RawMessage, status EventStatus) Event {
	return Event{
		ID:     id,
		Height: ht,
		Type:   typ,
		InTx:   inTx,
		Event:  evt,
		Status: status,
	}
}

// Empty determinate whether the event is empty
func (evt Event) Empty() bool {
	return evt.InTx.ID.IsEmpty()
}

// Events is a slice of events
type Events []Event

// PopByInHash Pops an event out of the event list by hash ID
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

// EventSwap event for swap action
type EventSwap struct {
	Pool         common.Asset `json:"pool"`
	PriceTarget  sdk.Uint     `json:"price_target"`
	TradeSlip    sdk.Uint     `json:"trade_slip"`
	LiquidityFee sdk.Uint     `json:"liquidity_fee"`
}

// NewEventSwap create a new swap event
func NewEventSwap(pool common.Asset, priceTarget, fee, tradeSlip sdk.Uint) EventSwap {
	return EventSwap{
		Pool:         pool,
		PriceTarget:  priceTarget,
		TradeSlip:    tradeSlip,
		LiquidityFee: fee,
	}
}

// Type return a string that represent the type, it should not duplicated with other event
func (e EventSwap) Type() string {
	return SwapEventType
}

// EventStake stake event
type EventStake struct {
	Pool       common.Asset `json:"pool"`
	StakeUnits sdk.Uint     `json:"stake_units"`
}

// NewEventStake create a new stake event
func NewEventStake(pool common.Asset, su sdk.Uint) EventStake {
	return EventStake{
		Pool:       pool,
		StakeUnits: su,
	}
}

// Type return the event type
func (e EventStake) Type() string {
	return StakeEventType
}

// EventUnstake represent unstake
type EventUnstake struct {
	Pool        common.Asset `json:"pool"`
	StakeUnits  sdk.Uint     `json:"stake_units"`
	BasisPoints int64        `json:"basis_points"` // 1 ==> 10,0000
	Asymmetry   sdk.Dec      `json:"asymmetry"`    // -1.0 <==> 1.0
}

// NewEventUnstake create a new unstake event
func NewEventUnstake(pool common.Asset, su sdk.Uint, basisPts int64, asym sdk.Dec) EventUnstake {
	return EventUnstake{
		Pool:        pool,
		StakeUnits:  su,
		BasisPoints: basisPts,
		Asymmetry:   asym,
	}
}

// Type return the unstake event type
func (e EventUnstake) Type() string {
	return UnstakeEventType
}

// EventAdminConfig represent admin config change events
type EventAdminConfig struct {
	Key   string
	Value string
}

// NewEventAdminConfig create a new admin config event
func NewEventAdminConfig(key, value string) EventAdminConfig {
	return EventAdminConfig{
		Key:   key,
		Value: value,
	}
}

// Type return the type of admin config event
func (e EventAdminConfig) Type() string {
	return AdminConfigEventType
}

// EventAdd represent add operation
type EventAdd struct {
	Pool common.Asset `json:"pool"`
}

// NewEventAdd create a new add event
func NewEventAdd(pool common.Asset) EventAdd {
	return EventAdd{
		Pool: pool,
	}
}

// Type return add event type
func (e EventAdd) Type() string {
	return AddEventType
}

// EventPool represent pool change event
type EventPool struct {
	Pool   common.Asset `json:"pool"`
	Status PoolStatus   `json:"status"`
}

// NewEventPool create a new pool change event
func NewEventPool(pool common.Asset, status PoolStatus) EventPool {
	return EventPool{
		Pool:   pool,
		Status: status,
	}
}

// Type return pool event type
func (e EventPool) Type() string {
	return PoolEventType
}

// PoolAmt pool asset amount
type PoolAmt struct {
	Asset  common.Asset `json:"asset"`
	Amount int64        `json:"amount"`
}

// EventRewards reward event
type EventRewards struct {
	BondReward  sdk.Uint  `json:"bond_reward"`
	PoolRewards []PoolAmt `json:"pool_rewards"`
}

// NewEventRewards create a new reward event
func NewEventRewards(bondReward sdk.Uint, poolRewards []PoolAmt) EventRewards {
	return EventRewards{
		BondReward:  bondReward,
		PoolRewards: poolRewards,
	}
}

// Type return reward event type
func (e EventRewards) Type() string {
	return RewardEventType
}

// EmptyRefundEvent represent refund
type EmptyRefundEvent struct {
}

// NewEmptyRefundEvent create a new EmptyRefundEvent
func NewEmptyRefundEvent() EmptyRefundEvent {
	return EmptyRefundEvent{}
}

// Type return EmptyRefundEvent type
func (e EmptyRefundEvent) Type() string {
	return EmptyRefundEventType
}
