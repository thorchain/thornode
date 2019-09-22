package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgAddTrustAccount defines a MsgAddTrustAccount message
type MsgAddTrustAccount struct {
	TrustAccount TrustAccount   `json:"trust_account"`
	Signer       sdk.AccAddress `json:"signer"`
}

// NewMsgAddTrustAccount is a constructor function for NewMsgAddTrustAccount
func NewMsgAddTrustAccount(trust TrustAccount, signer sdk.AccAddress) MsgAddTrustAccount {
	return MsgAddTrustAccount{
		TrustAccount: trust,
		Signer:       signer,
	}
}

// Route should return the cmname of the module
func (msg MsgAddTrustAccount) Route() string { return RouterKey }

// Type should return the action
func (msg MsgAddTrustAccount) Type() string { return "set_trust_account" }

// ValidateBasic runs stateless checks on the message
func (msg MsgAddTrustAccount) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if err := msg.TrustAccount.IsValid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgAddTrustAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgAddTrustAccount) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
