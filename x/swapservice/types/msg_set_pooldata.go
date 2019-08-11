package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSetPoolData defines a SetPoolData message
// We keep this for now , as a mechanism to set up a new pool when it is not in the genesis file
// the pool changes when stake / swap happens
type MsgSetPoolData struct {
	BalanceRune  Amount         `json:"balance_rune"`  // balance rune
	BalanceToken Amount         `json:"balance_token"` // balance of token
	Ticker       Ticker         `json:"ticker"`        // Ticker means the token symbol
	Status       PoolStatus     `json:"status"`        // pool status
	Owner        sdk.AccAddress `json:"owner"`
}

// NewMsgSetPoolData is a constructor function for MsgSetPoolData
func NewMsgSetPoolData(ticker Ticker, status PoolStatus, owner sdk.AccAddress) MsgSetPoolData {
	return MsgSetPoolData{
		Ticker:       ticker,
		BalanceRune:  ZeroAmount,
		BalanceToken: ZeroAmount,
		Status:       status,
		Owner:        owner,
	}
}

// Route should return the pooldata of the module
func (msg MsgSetPoolData) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetPoolData) Type() string { return "set_pooldata" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetPoolData) ValidateBasic() sdk.Error {
	if msg.Ticker.Empty() {
		return sdk.ErrUnknownRequest("Pool Ticker cannot be empty")
	}
	if msg.BalanceRune.Empty() {
		return sdk.ErrUnknownRequest("Rune balance cannot be empty")
	}
	if msg.BalanceToken.Empty() {
		return sdk.ErrUnknownRequest("Token balance cannot be empty")
	}
	if err := msg.Status.Valid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
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
