package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSetStakeData defines a SetStakeData message
type MsgSetStakeData struct {
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
		Name:   name,
		Ticker: ticker,
		Token:  token,
		Rune:   r,
		Owner:  owner,
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
