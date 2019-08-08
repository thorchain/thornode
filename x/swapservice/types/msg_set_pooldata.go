package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSetPoolData defines a SetPoolData message
// We keep this for now , as a mechanism to set up a new pool when it is not in the genesis file
// the pool changes when stake / swap happens
type MsgSetPoolData struct {
	BalanceRune  string         `json:"balance_rune"`  // balance rune
	BalanceToken string         `json:"balance_token"` // balance of token
	Ticker       string         `json:"ticker"`        // Ticker means the token symbol
	TokenName    string         `json:"token_name"`    // usually it is a more user friendly token name
	PoolAddress  string         `json:"pool_address"`  // Pool Address on binance chain
	Status       PoolStatus     `json:"status"`        // pool status
	Owner        sdk.AccAddress `json:"owner"`
}

// NewMsgSetPoolData is a constructor function for MsgSetPoolData
func NewMsgSetPoolData(tokenName, ticker string, poolAddress string, status PoolStatus, owner sdk.AccAddress) MsgSetPoolData {
	return MsgSetPoolData{
		Ticker:       strings.ToUpper(ticker),
		TokenName:    tokenName,
		BalanceRune:  "0",
		BalanceToken: "0",
		PoolAddress:  poolAddress,
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
	if len(msg.Ticker) == 0 {
		return sdk.ErrUnknownRequest("Pool Ticker cannot be empty")
	}
	if len(msg.TokenName) == 0 {
		return sdk.ErrUnknownRequest("Pool TokenName cannot be empty")
	}
	if len(msg.PoolAddress) == 0 {
		return sdk.ErrUnknownRequest("pool address can't be empty")
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
