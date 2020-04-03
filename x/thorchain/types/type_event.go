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
	Fee    common.Fee      `json:"fee"`
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
	RefundEventType      = `refund`
	BondEventType        = `bond`
	GasEventType         = `gas`
	ReserveEventType     = `reserve`
	SlashEventType       = `slash`
)

// NewEvent create a new  event
func NewEvent(typ string, ht int64, inTx common.Tx, evt json.RawMessage, status EventStatus) Event {
	return Event{
		Height: ht,
		Type:   typ,
		InTx:   inTx,
		Event:  evt,
		Status: status,
		Fee: common.Fee{
			Coins:      common.Coins{},
			PoolDeduct: sdk.ZeroUint(),
		},
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
	Pool               common.Asset `json:"pool"`
	PriceTarget        sdk.Uint     `json:"price_target"`
	TradeSlip          sdk.Uint     `json:"trade_slip"`
	LiquidityFee       sdk.Uint     `json:"liquidity_fee"`
	LiquidityFeeInRune sdk.Uint     `json:"liquidity_fee_in_rune"`
}

// NewEventSwap create a new swap event
func NewEventSwap(pool common.Asset, priceTarget, fee, tradeSlip, liquidityFeeInRune sdk.Uint) EventSwap {
	return EventSwap{
		Pool:               pool,
		PriceTarget:        priceTarget,
		TradeSlip:          tradeSlip,
		LiquidityFee:       fee,
		LiquidityFeeInRune: liquidityFeeInRune,
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

// NewEventRefund create a new EventRefund
func NewEventRefund(code sdk.CodeType, reason string) EventRefund {
	return EventRefund{
		Code:   code,
		Reason: reason,
	}
}

// EventRefund represent a refund activity , and contains the reason why it get refund
type EventRefund struct {
	Code   sdk.CodeType `json:"code"`
	Reason string       `json:"reason"`
}

// Type return reward event type
func (e EventRefund) Type() string {
	return RefundEventType
}

type BondType string

const (
	BondPaid     BondType = `bond_paid`
	BondReturned BondType = `bond_returned`
)

// EventBond bond paid or returned event
type EventBond struct {
	Amount   sdk.Uint `json:"amount"`
	BondType BondType `json:"bond_type"`
}

// Type return bond event Type
func (e EventBond) Type() string {
	return BondEventType
}

// NewEventBond create a new Bond Event
func NewEventBond(amount sdk.Uint, bondType BondType) EventBond {
	return EventBond{
		Amount:   amount,
		BondType: bondType,
	}
}

type GasType string

type GasPool struct {
	Asset    common.Asset `json:"asset"`
	AssetAmt sdk.Uint     `json:"asset_amt"`
	RuneAmt  sdk.Uint     `json:"rune_amt"`
}

// EventGas represent the events happened in thorchain related to Gas
type EventGas struct {
	Pools []GasPool `json:"pools"`
}

// NewEventGas create a new EventGas instance
func NewEventGas() *EventGas {
	return &EventGas{
		Pools: []GasPool{},
	}
}

// UpsertGasPool update the Gas Pools hold by EventGas instance
// if the given gasPool already exist, then it merge the gasPool with internal one , otherwise add it to the list
func (e *EventGas) UpsertGasPool(pool GasPool) {
	for i, p := range e.Pools {
		if p.Asset == pool.Asset {
			e.Pools[i].RuneAmt = p.RuneAmt.Add(pool.RuneAmt)
			e.Pools[i].AssetAmt = p.AssetAmt.Add(pool.AssetAmt)
			return
		}
	}
	e.Pools = append(e.Pools, pool)
}

// Type return event type
func (e *EventGas) Type() string {
	return GasEventType
}

// EventReserve Reserve event type
type EventReserve struct {
	ReserveContributor ReserveContributor `json:"reserve_contributor"`
}

// NewEventReserve create a new instance of EventReserve
func NewEventReserve(contributor ReserveContributor) EventReserve {
	return EventReserve{
		ReserveContributor: contributor,
	}
}

func (e EventReserve) Type() string {
	return ReserveEventType
}

// EventSlash represent a change in pool balance which caused by slash a node account
type EventSlash struct {
	Pool        common.Asset `json:"pool"`
	SlashAmount []PoolAmt    `json:"slash_amount"`
}

func NewEventSlash(pool common.Asset, slashAmount []PoolAmt) EventSlash {
	return EventSlash{
		Pool:        pool,
		SlashAmount: slashAmount,
	}
}

// Type return slash event type
func (e EventSlash) Type() string {
	return SlashEventType
}
