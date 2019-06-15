package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const RouterKey = ModuleName // this was defined in your key.go file

// MsgSetPoolData defines a SetPoolData message
type MsgSetPoolData struct {
	PoolID       string `json:"pool_id"`
	BalanceAtom  string `json:"balance_atom"`
	BalanceToken string `json:"balance_token"`
	Ticker       string `json:"ticker"`
	TokenName    string `json:"token_name"`
}

// NewMsgSetPoolData is a constructor function for MsgSetPoolData
func NewMsgSetPoolData(tokenName, ticker string) MsgSetPoolData {
	return MsgSetPoolData{
		Ticker:    strings.ToUpper(ticker),
		TokenName: tokenName,
	}
}

// Route should return the pooldata of the module
func (msg MsgSetPoolData) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetPoolData) Type() string { return "set_pooldata" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetPoolData) ValidateBasic() sdk.Error {
	if len(msg.Ticker) == 0 {
		return sdk.ErrUnknownRequest("Pool Ticker cannot be empty")
	}
	if len(msg.TokenName) == 0 {
		return sdk.ErrUnknownRequest("Pool TokenName cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetPoolData) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetPoolData) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}
