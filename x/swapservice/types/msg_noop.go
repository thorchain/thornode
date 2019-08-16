package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgNoOp defines a no op message
type MsgNoOp struct {
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgNoOp is a constructor function for MsgNoOp
func NewMsgNoOp(signer sdk.AccAddress) MsgNoOp {
	return MsgNoOp{
		Signer: signer,
	}
}

// Route should return the pooldata of the module
func (msg MsgNoOp) Route() string { return RouterKey }

// Type should return the action
func (msg MsgNoOp) Type() string { return "set_noop" }

// ValidateBasic runs stateless checks on the message
func (msg MsgNoOp) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgNoOp) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgNoOp) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
