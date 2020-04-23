package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgBan defines a MsgBan message
type MsgBan struct {
	NodeAddress sdk.AccAddress `json:"node_address"`
	Signer      sdk.AccAddress `json:"signer"`
}

// NewMsgBan is a constructor function for NewMsgBan
func NewMsgBan(addr, signer sdk.AccAddress) MsgBan {
	return MsgBan{
		NodeAddress: addr,
		Signer:      signer,
	}
}

// Route should return the cmname of the module
func (msg MsgBan) Route() string { return RouterKey }

// Type should return the action
func (msg MsgBan) Type() string { return "ban" }

// ValidateBasic runs stateless checks on the message
func (msg MsgBan) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if msg.NodeAddress.Empty() {
		return sdk.ErrInvalidAddress(msg.NodeAddress.String())
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgBan) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgBan) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
