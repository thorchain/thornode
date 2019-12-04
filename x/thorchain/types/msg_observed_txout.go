package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

// MsgObservedTxOut defines a MsgObservedTxOut message
type MsgObservedTxOut struct {
	Txs    []common.Tx    `json:"tx_in"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgObservedTxOut is a constructor function for MsgObservedTxOut
func NewMsgObservedTxOut(txs []common.Tx, signer sdk.AccAddress) MsgObservedTxOut {
	return MsgObservedTxOut{
		Txs:    txs,
		Signer: signer,
	}
}

// Route should return the cmname of the module
func (msg MsgObservedTxOut) Route() string { return RouterKey }

// Type should return the action
func (msg MsgObservedTxOut) Type() string { return "set_observed_txout" }

// ValidateBasic runs stateless checks on the message
func (msg MsgObservedTxOut) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if len(msg.Txs) == 0 {
		return sdk.ErrUnknownRequest("Txs cannot be empty")
	}
	for _, tx := range msg.Txs {
		if err := tx.IsValid(); err != nil {
			return sdk.ErrUnknownRequest(err.Error())
		}
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgObservedTxOut) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgObservedTxOut) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
