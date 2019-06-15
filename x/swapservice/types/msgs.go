package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const RouterKey = ModuleName // this was defined in your key.go file

// MsgSetPoolData defines a SetPoolData message
type MsgSetPoolData struct {
	PoolID       string         `json:"pool_id"`
	BalanceAtom  string         `json:"balance_atom"`
	BalanceToken string         `json:"balance_token"`
	Ticker       string         `json:"ticker"`
	TokenName    string         `json:"token_name"`
	Owner        sdk.AccAddress `json:"owner"`
}

// NewMsgSetPoolData is a constructor function for MsgSetPoolData
func NewMsgSetPoolData(tokenName, ticker string, owner sdk.AccAddress) MsgSetPoolData {
	return MsgSetPoolData{
		PoolID:    fmt.Sprintf("pool-%s", strings.ToUpper(ticker)),
		Ticker:    strings.ToUpper(ticker),
		TokenName: tokenName,
		Owner:     owner,
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
	return []sdk.AccAddress{msg.Owner}
}

// MsgSetAccData defines a SetAccData message
type MsgSetAccData struct {
	AccID string         `json:"acc_id"`
	Name  string         `json:"name"`
	ATOM  string         `json:"atom"`
	BTC   string         `json:"btc"`
	Owner sdk.AccAddress `json:"owner"`
}

// NewMsgSetPoolData is a constructor function for MsgSetPoolData
func NewMsgSetAccData(name, atom, btc string, owner sdk.AccAddress) MsgSetAccData {
	return MsgSetAccData{
		AccID: fmt.Sprintf("acc-%s", strings.ToLower(name)),
		Name:  strings.ToLower(name),
		ATOM:  atom,
		BTC:   btc,
		Owner: owner,
	}
}

// Route should return the pooldata of the module
func (msg MsgSetAccData) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetAccData) Type() string { return "set_accdata" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetAccData) ValidateBasic() sdk.Error {
	if len(msg.Name) == 0 {
		return sdk.ErrUnknownRequest("Account Name cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetAccData) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetAccData) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

// MsgSetStakeData defines a SetStakeData message
type MsgSetStakeData struct {
	StakeID string         `json:"acc_id"`
	Name    string         `json:"name"`
	Atom    string         `json:"atom"`
	Token   string         `json:"token"`
	Owner   sdk.AccAddress `json:"owner"`
}

// NewMsgSetPoolData is a constructor function for MsgSetPoolData
func NewMsgSetStakeData(ticker, name, atom, token string, owner sdk.AccAddress) MsgSetStakeData {
	return MsgSetStakeData{
		StakeID: fmt.Sprintf("stake-%s", strings.ToUpper(ticker)),
		Name:    name,
		Atom:    atom,
		Token:   token,
		Owner:   owner,
	}
}

// Route should return the pooldata of the module
func (msg MsgSetStakeData) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetStakeData) Type() string { return "set_stakedata" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetStakeData) ValidateBasic() sdk.Error {
	if len(msg.Name) == 0 {
		return sdk.ErrUnknownRequest("Stake Name cannot be empty")
	}
	if len(msg.Atom) == 0 {
		return sdk.ErrUnknownRequest("Stake Atom cannot be empty")
	}
	if len(msg.Token) == 0 {
		return sdk.ErrUnknownRequest("Stake Token cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetStakeData) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetStakeData) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}
