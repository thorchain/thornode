package thorchain

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// EmitGasEvents will emit the gas change in a block
func EmitGasEvents(ctx sdk.Context, keeper Keeper) error {
	defer keeper.RemoveBlockGas(ctx)
	blockGas, err := keeper.GetBlockGas(ctx)
	if err != nil {
		return err
	}
	// nothing
	if blockGas.IsEmpty() {
		return nil
	}
	gasEvent := NewEventGas(blockGas.GasSpend, blockGas.GasTopup, blockGas.GasReimburse)
	buf, err := json.Marshal(gasEvent)
	if err != nil {
		return err
	}
	evt := NewEvent(gasEvent.Type(), ctx.BlockHeight(),
		common.Tx{ID: common.BlankTxID},
		buf, EventSuccess)
	return keeper.UpsertEvent(ctx, evt)
}
