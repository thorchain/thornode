package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const RouterKey = ModuleName // this was defined in your key.go file

// MsgSetPoolData defines a SetPoolData message
type MsgSetPoolData struct {
	PoolID       string         `json:"pool_id"`       // generated automatically based on the ticker
	BalanceRune  string         `json:"balance_rune"`  // balance rune
	BalanceToken string         `json:"balance_token"` // balance of token
	Ticker       string         `json:"ticker"`        // Ticker means the token symbol
	TokenName    string         `json:"token_name"`    // usually it is a more user friendly token name
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
	AccID  string         `json:"acc_id"`
	Name   string         `json:"name"`
	Ticker string         `json:"ticker"`
	Amount string         `json:"token"`
	Owner  sdk.AccAddress `json:"owner"`
}

// NewMsgSetAccData is a constructor function for MsgSetAccData
func NewMsgSetAccData(name, ticker, amount string, owner sdk.AccAddress) MsgSetAccData {
	return MsgSetAccData{
		AccID:  fmt.Sprintf("acc-%s", strings.ToLower(name)),
		Name:   strings.ToLower(name),
		Ticker: ticker,
		Amount: amount,
		Owner:  owner,
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
	StakeID string         `json:"stake_id"` // StakeID generated automatically based on the ticker
	Name    string         `json:"name"`     // token usually is a more user friendly name
	Ticker  string         `json:"ticker"`   // ticker means the symbol
	Token   string         `json:"token"`    // the amount of token stake
	Rune    string         `json:"rune"`     // the amount of rune stake
	Owner   sdk.AccAddress `json:"owner"`
}

// NewMsgSetStakeData is a constructor function for MsgSetStakeData
func NewMsgSetStakeData(name, ticker, r, token string, owner sdk.AccAddress) MsgSetStakeData {
	return MsgSetStakeData{
		StakeID: fmt.Sprintf("stake-%s", strings.ToUpper(ticker)),
		Name:    name,
		Ticker:  ticker,
		Token:   token,
		Rune:    r,
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
	if len(msg.Ticker) == 0 {
		return sdk.ErrUnknownRequest("Stake Ticker cannot be empty")
	}
	if len(msg.Rune) == 0 {
		return sdk.ErrUnknownRequest("Stake Rune cannot be empty")
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

// MsgSwap defines a MsgSwap message
type MsgSwap struct {
	PoolID       string         `json:"pool_id"`
	SourceTicker string         `json:"source_ticker"`
	TargetTicker string         `json:"target_ticker"`
	Requester    string         `json:"requester"`
	Destination  string         `json:"destination"`
	Amount       string         `json:"amount"`
	Owner        sdk.AccAddress `json:"owner"`
}

// NewMsgSwap is a constructor function for MsgSwap
func NewMsgSwap(source, target, amount, requester, destination string, owner sdk.AccAddress) MsgSwap {
	return MsgSwap{
		SourceTicker: source,
		TargetTicker: target,
		Amount:       amount,
		Requester:    requester,
		Destination:  destination,
		Owner:        owner,
	}
}

// Route should return the pooldata of the module
func (msg MsgSwap) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSwap) Type() string { return "set_swap" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSwap) ValidateBasic() sdk.Error {
	if len(msg.SourceTicker) == 0 {
		return sdk.ErrUnknownRequest("Swap Source Ticker cannot be empty")
	}
	if len(msg.TargetTicker) == 0 {
		return sdk.ErrUnknownRequest("Swap Target cannot be empty")
	}
	if len(msg.Amount) == 0 {
		return sdk.ErrUnknownRequest("Swap Amount cannot be empty")
	}
	if len(msg.Requester) == 0 {
		return sdk.ErrUnknownRequest("Swap Requester cannot be empty")
	}
	if len(msg.Destination) == 0 {
		return sdk.ErrUnknownRequest("Swap Destination cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSwap) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSwap) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

// MsgStake defines a MsgStake message
type MsgStake struct {
	Name        string         `json:"name"`
	Ticker      string         `json:"ticker"`
	AtomAmount  string         `json:"atom_amount"`
	TokenAmount string         `json:"token_amount"`
	Owner       sdk.AccAddress `json:"owner"`
}
