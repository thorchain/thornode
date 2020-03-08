package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgRagnarok defines a MsgRagnarok message
type MsgRagnarok struct {
	Tx          ObservedTx     `json:"tx"`
	BlockHeight int64          `json:"block_height"`
	Signer      sdk.AccAddress `json:"signer"`
}

// NewMsgRagnarok is a constructor function for MsgRagnarok
func NewMsgRagnarok(tx ObservedTx, blockHeight int64, signer sdk.AccAddress) MsgRagnarok {
	return MsgRagnarok{
		Tx:          tx,
		BlockHeight: blockHeight,
		Signer:      signer,
	}
}

// Route should return the name of the module
func (msg MsgRagnarok) Route() string { return RouterKey }

// Type should return the action
func (msg MsgRagnarok) Type() string { return "ragnarok" }

// ValidateBasic runs stateless checks on the message
func (msg MsgRagnarok) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if msg.BlockHeight <= 0 {
		return sdk.ErrUnknownRequest("invalid block height")
	}
	if err := msg.Tx.Valid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgRagnarok) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgRagnarok) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
