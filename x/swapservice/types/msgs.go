package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const RouterKey = ModuleName // this was defined in your key.go file

// MsgSetAccData defines a SetAccData message
// TODO: given we don't hold asset ,thus we don't need to set Account Data
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
	StakeID       string         `json:"stake_id"`       // StakeID generated automatically based on the ticker
	Name          string         `json:"name"`           // token usually is a more user friendly name
	Ticker        string         `json:"ticker"`         // ticker means the symbol
	Token         string         `json:"token"`          // the amount of token stake
	Rune          string         `json:"rune"`           // the amount of rune stake
	PublicAddress string         `json:"public_address"` // Staker's address on binance chain
	Owner         sdk.AccAddress `json:"owner"`
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

// MsgStake defines a MsgStake message
// TODO don't need it,and it is not used either , for remove
type MsgStake struct {
	Name        string         `json:"name"`
	Ticker      string         `json:"ticker"`
	AtomAmount  string         `json:"atom_amount"`
	TokenAmount string         `json:"token_amount"`
	Owner       sdk.AccAddress `json:"owner"`
}
