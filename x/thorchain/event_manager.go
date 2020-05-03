package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// EventManager define methods need to be support to manage events
type EventManager interface {
	BeginBlock(ctx sdk.Context)
	EndBlock(ctx sdk.Context, keeper Keeper)
	GetBlockEvents(ctx sdk.Context, keeper Keeper, height int64) (*BlockEvents, error)
	CompleteEvents(ctx sdk.Context, keeper Keeper, height int64, txID common.TxID, txs common.Txs, eventStatus EventStatus) error
	FailStalePendingEvents(ctx sdk.Context, constantValues constants.ConstantValues, keeper Keeper)
	UpdateEventFee(ctx sdk.Context, txID common.TxID, fee common.Fee)
	AddEvent(event Event)
}

// EventMgr implement EventManager interface
type EventMgr struct {
	blockEvents *BlockEvents
}

// NewEventMgr create a new instance of EventMgr
func NewEventMgr() *EventMgr {
	return &EventMgr{}
}

// BeginBlock is going to create a brand new BlockEvents
func (m *EventMgr) BeginBlock(ctx sdk.Context) {
	m.blockEvents = NewBlockEvents(ctx.BlockHeight())
}

// EndBlock will write the block event to storage
func (m *EventMgr) EndBlock(ctx sdk.Context, keeper Keeper) {
	keeper.SetBlockEvents(ctx, m.blockEvents)
}

// GetBlockEvents return the instance of block events on the given height
func (m *EventMgr) GetBlockEvents(ctx sdk.Context, keeper Keeper, height int64) (*BlockEvents, error) {
	return keeper.GetBlockEvents(ctx, height)
}

// CompleteEvents Mark an event in the given block height to the given status
func (m *EventMgr) CompleteEvents(ctx sdk.Context, keeper Keeper, height int64, txID common.TxID, txs common.Txs, eventStatus EventStatus) error {
	ctx.Logger().Info(fmt.Sprintf("txid(%s)", txID))
	blockEvent, err := keeper.GetBlockEvents(ctx, height)
	if err != nil {
		return fmt.Errorf("fail to get block event on height(%d): %w", height, err)
	}
	for idx, e := range blockEvent.Events {
		if !e.InTx.ID.Equals(txID) {
			continue
		}
		if e.Status == EventSuccess {
			continue
		}
		ctx.Logger().Info(fmt.Sprintf("set event to %s,txid (%s) , txs: %s", eventStatus, txID, txs))
		outTxs := append(e.OutTxs, txs...)
		for i := 0; i < len(outTxs); i++ {
			duplicate := false
			for j := i + 1; j < len(outTxs); j++ {
				if outTxs[i].Equals(outTxs[j]) {
					duplicate = true
				}
			}
			if !duplicate {
				blockEvent.Events[idx].OutTxs = append(blockEvent.Events[idx].OutTxs, outTxs[i])
			}
		}
		if eventStatus == EventRefund {
			// we need to check we refunded all the coins that need to be refunded from in tx
			// before updating status to complete, we use the count of voter actions to check
			voter, err := keeper.GetObservedTxVoter(ctx, e.InTx.ID)
			if err != nil {
				return fmt.Errorf("fail to get observed tx voter: %w", err)
			}
			if len(voter.Actions) == len(blockEvent.Events[idx].OutTxs) {
				blockEvent.Events[idx].Status = eventStatus
			}
		} else {
			blockEvent.Events[idx].Status = eventStatus
		}
	}
	// save the changes
	keeper.SetBlockEvents(ctx, blockEvent)
	return nil
}

// UpdateEventFee
func (m *EventMgr) UpdateEventFee(ctx sdk.Context, txID common.TxID, fee common.Fee) {
	ctx.Logger().Info("update event fee txid", "tx", txID.String())
	var swapEventIdxes []int
	for idx, e := range m.blockEvents.Events {
		if !e.InTx.ID.Equals(txID) {
			continue
		}
		// special treatment for swap events, as double swap will emit two events, which share the same txID,Only the second one has fee
		if e.Type != SwapEventType {
			ctx.Logger().Info(fmt.Sprintf("Update fee for event %s, fee:%s", txID, fee))
			m.blockEvents.Events[idx].Fee.Coins = append(m.blockEvents.Events[idx].Fee.Coins, fee.Coins...)
			m.blockEvents.Events[idx].Fee.PoolDeduct = m.blockEvents.Events[idx].Fee.PoolDeduct.Add(fee.PoolDeduct)
			continue
		}
		swapEventIdxes = append(swapEventIdxes, idx)
	}
	if len(swapEventIdxes) == 0 {
		return
	}
	ctx.Logger().Info(fmt.Sprintf("update fee for event %s, fee: %s", txID, fee))
	idx := swapEventIdxes[len(swapEventIdxes)-1]
	m.blockEvents.Events[idx].Fee.Coins = append(m.blockEvents.Events[idx].Fee.Coins, fee.Coins...)
	m.blockEvents.Events[idx].Fee.PoolDeduct = m.blockEvents.Events[idx].Fee.PoolDeduct.Add(fee.PoolDeduct)
}

// AddEvent add an event to block event
func (m *EventMgr) AddEvent(event Event) {
	m.blockEvents.AddEvent(event)
}

func (m *EventMgr) FailStalePendingEvents(ctx sdk.Context, constantValues constants.ConstantValues, keeper Keeper) {
	// fail stale pending events
	signingTransPeriod := constantValues.GetInt64Value(constants.SigningTransactionPeriod)
	targetBlockHeight := ctx.BlockHeight() - 2*signingTransPeriod
	if targetBlockHeight < 0 {
		return
	}
	blockEvent, err := m.GetBlockEvents(ctx, keeper, targetBlockHeight)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("fail to get block events on height: %d", targetBlockHeight), "error", err)
		return
	}
	if blockEvent == nil {
		return
	}
	for idx, e := range blockEvent.Events {
		if e.Status == EventPending {
			blockEvent.Events[idx].Status = EventFail
		}
	}
	keeper.SetBlockEvents(ctx, blockEvent)
}
