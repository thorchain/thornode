package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// MsgSetAdminConfig defines a MsgSetAdminConfig message
type MsgSetAdminConfig struct {
	AdminConfig AdminConfig    `json:"admin_config"`
	Signer      sdk.AccAddress `json:"signer"`
}

// NewMsgSetAdminConfig is a constructor function for MsgSetAdminConfig
func NewMsgSetAdminConfig(key, value string, signer sdk.AccAddress) MsgSetAdminConfig {
	return MsgSetAdminConfig{
		AdminConfig: NewAdminConfig(key, value),
		Signer:      signer,
	}
}

// Route should return the cmname of the module
func (msg MsgSetAdminConfig) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetAdminConfig) Type() string { return "set_admin_config" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetAdminConfig) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if err := msg.AdminConfig.Valid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetAdminConfig) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetAdminConfig) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
