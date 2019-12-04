package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

// MsgObservedTxIn defines a MsgObservedTxIn message
type MsgObservedTxIn struct {
	Txs    []common.Tx    `json:"tx_in"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgObservedTxIn is a constructor function for MsgObservedTxIn
func NewMsgObservedTxIn(txs []common.Tx, signer sdk.AccAddress) MsgObservedTxIn {
	return MsgObservedTxIn{
		Txs:    txs,
		Signer: signer,
	}
}

// Route should return the cmname of the module
func (msg MsgObservedTxIn) Route() string { return RouterKey }

// Type should return the action
func (msg MsgObservedTxIn) Type() string { return "set_observed_txin" }

// ValidateBasic runs stateless checks on the message
func (msg MsgObservedTxIn) ValidateBasic() sdk.Error {
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
func (msg MsgObservedTxIn) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgObservedTxIn) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
