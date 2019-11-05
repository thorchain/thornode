package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSetTrustAccount defines a MsgSetTrustAccount message
type MsgSetTrustAccount struct {
	TrustAccount TrustAccount   `json:"trust_account"`
	Signer       sdk.AccAddress `json:"signer"`
}

// NewMsgSetTrustAccount is a constructor function for NewMsgAddTrustAccount
func NewMsgSetTrustAccount(trustAccount TrustAccount, signer sdk.AccAddress) MsgSetTrustAccount {
	return MsgSetTrustAccount{
		TrustAccount: trustAccount,
		Signer:       signer,
	}
}

// Route should return the cmname of the module
func (msg MsgSetTrustAccount) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetTrustAccount) Type() string { return "set_trust_account" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetTrustAccount) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if err := msg.TrustAccount.IsValid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetTrustAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetTrustAccount) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
