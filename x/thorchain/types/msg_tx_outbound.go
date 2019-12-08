package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// MsgOutboundTx defines a MsgOutboundTx message
type MsgOutboundTx struct {
	Tx     ObservedTx     `json:"tx"`
	InTxID common.TxID    `json:"tx_id"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgOutboundTx is a constructor function for MsgOutboundTx
func NewMsgOutboundTx(tx ObservedTx, txID common.TxID, signer sdk.AccAddress) MsgOutboundTx {
	return MsgOutboundTx{
		Tx:     tx,
		InTxID: txID,
		Signer: signer,
	}
}

// Route should return the cmname of the module
func (msg MsgOutboundTx) Route() string { return RouterKey }

// Type should return the action
func (msg MsgOutboundTx) Type() string { return "set_tx_outbound" }

// ValidateBasic runs stateless checks on the message
func (msg MsgOutboundTx) ValidateBasic() sdk.Error {
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
func (msg MsgOutboundTx) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgOutboundTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
