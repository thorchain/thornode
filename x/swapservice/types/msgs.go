package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const RouterKey = ModuleName // this was defined in your key.go file

// MsgSetPoolData defines a SetPoolData message
type MsgSetPoolData struct {
	PoolData  string         `json:"pooldata"`
	Value string         `json:"value"`
	Owner sdk.AccAddress `json:"owner"`
}

// NewMsgSetPoolData is a constructor function for MsgSetPoolData
func NewMsgSetPoolData(pooldata string, value string, owner sdk.AccAddress) MsgSetPoolData {
	return MsgSetPoolData{
		PoolData:  pooldata,
		Value: value,
		Owner: owner,
	}
}

// Route should return the pooldata of the module
func (msg MsgSetPoolData) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetPoolData) Type() string { return "set_pooldata" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetPoolData) ValidateBasic() sdk.Error {
	if msg.Owner.Empty() {
		return sdk.ErrInvalidAddress(msg.Owner.String())
	}
	if len(msg.PoolData) == 0 || len(msg.Value) == 0 {
		return sdk.ErrUnknownRequest("PoolData and/or Value cannot be empty")
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

// MsgBuyPoolData defines the BuyPoolData message
type MsgBuyPoolData struct {
	PoolData  string         `json:"pooldata"`
	Bid   sdk.Coins      `json:"bid"`
	Buyer sdk.AccAddress `json:"buyer"`
}

// NewMsgBuyPoolData is the constructor function for MsgBuyPoolData
func NewMsgBuyPoolData(pooldata string, bid sdk.Coins, buyer sdk.AccAddress) MsgBuyPoolData {
	return MsgBuyPoolData{
		PoolData:  pooldata,
		Bid:   bid,
		Buyer: buyer,
	}
}

// Route should return the pooldata of the module
func (msg MsgBuyPoolData) Route() string { return RouterKey }

// Type should return the action
func (msg MsgBuyPoolData) Type() string { return "buy_pooldata" }

// ValidateBasic runs stateless checks on the message
func (msg MsgBuyPoolData) ValidateBasic() sdk.Error {
	if msg.Buyer.Empty() {
		return sdk.ErrInvalidAddress(msg.Buyer.String())
	}
	if len(msg.PoolData) == 0 {
		return sdk.ErrUnknownRequest("PoolData cannot be empty")
	}
	if !msg.Bid.IsAllPositive() {
		return sdk.ErrInsufficientCoins("Bids must be positive")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgBuyPoolData) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgBuyPoolData) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Buyer}
}
