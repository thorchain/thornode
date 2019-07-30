package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSwap defines a MsgSwap message
type MsgSwap struct {
	SourceTicker string         `json:"source_ticker"` // source token
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
