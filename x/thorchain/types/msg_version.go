package types

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSetVersion defines a MsgSetVersion message
type MsgSetVersion struct {
	Version semver.Version `json:"version"`
	Signer  sdk.AccAddress `json:"signer"`
}

// NewMsgSetVersion is a constructor function for NewMsgSetVersion
func NewMsgSetVersion(version semver.Version, signer sdk.AccAddress) MsgSetVersion {
	return MsgSetVersion{
		Version: version,
		Signer:  signer,
	}
}

// Route should return the cmname of the module
func (msg MsgSetVersion) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetVersion) Type() string { return "set_version" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetVersion) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if err := msg.Version.Validate(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetVersion) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetVersion) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
