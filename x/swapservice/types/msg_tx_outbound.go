package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
)

// MsgOutboundTx defines a MsgOutboundTx message
type MsgOutboundTx struct {
	Height int64             `json:"height"`
	TxID   common.TxID       `json:"tx_id"`
	Sender common.BnbAddress `json:"sender"`
	Signer sdk.AccAddress    `json:"signer"`
}

// NewMsgOutboundTx is a constructor function for MsgOutboundTx
func NewMsgOutboundTx(txID common.TxID, height int64, sender common.BnbAddress, signer sdk.AccAddress) MsgOutboundTx {
	return MsgOutboundTx{
		Sender: sender,
		TxID:   txID,
		Height: height,
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
	if msg.Height <= 0 {
		return sdk.ErrUnknownRequest("Height must be above zero")
	}
	if msg.Sender.IsEmpty() {
		return sdk.ErrUnknownRequest("Sender cannot be empty")
	}
	if msg.TxID.IsEmpty() {
		return sdk.ErrUnknownRequest("TxID cannot be empty")
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
