package types

import (
	"net"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSetIPAddress defines a MsgSetIPAddress message
type MsgSetIPAddress struct {
	IPAddress string         `json:"ip_address"`
	Signer    sdk.AccAddress `json:"signer"`
}

// NewMsgSetIPAddress is a constructor function for NewMsgSetIPAddress
func NewMsgSetIPAddress(ip string, signer sdk.AccAddress) MsgSetIPAddress {
	return MsgSetIPAddress{
		IPAddress: ip,
		Signer:    signer,
	}
}

// Route should return the cmname of the module
func (msg MsgSetIPAddress) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetIPAddress) Type() string { return "set_ip_address" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetIPAddress) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if net.ParseIP(msg.IPAddress) == nil {
		return sdk.ErrUnknownRequest("invalid IP address")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetIPAddress) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetIPAddress) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
