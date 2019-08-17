package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MockTxOutSetter
type MockTxOutSetter struct {
}

func (mts MockTxOutSetter) SetTxOut(ctx sdk.Context, out *TxOut) {

}
