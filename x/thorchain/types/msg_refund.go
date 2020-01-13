package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// MsgRefundTx defines a MsgRefundTx message
type MsgRefundTx struct {
	Tx     ObservedTx     `json:"tx"`
	InTxID common.TxID    `json:"tx_id"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgRefundTx is a constructor function for MsgOutboundTx
func NewMsgRefundTx(tx ObservedTx, txID common.TxID, signer sdk.AccAddress) MsgRefundTx {
	return MsgRefundTx{
		Tx:     tx,
		InTxID: txID,
		Signer: signer,
	}
}

// Route should return the cmname of the module
func (msg MsgRefundTx) Route() string { return RouterKey }

// Type should return the action
func (msg MsgRefundTx) Type() string { return "set_tx_refund" }

// ValidateBasic runs stateless checks on the message
func (msg MsgRefundTx) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if msg.InTxID.IsEmpty() {
		return sdk.ErrUnknownRequest("In Tx ID cannot be empty")
	}
	if err := msg.Tx.Valid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgRefundTx) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgRefundTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
