package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSwap defines a MsgSwap message
type MsgSwap struct {
	RequestTxHash string         `json:"request_tx_hash"` // Request transaction hash on Binance chain
	SourceTicker  string         `json:"source_ticker"`   // source token
	TargetTicker  string         `json:"target_ticker"`   // target token
	Requester     string         `json:"requester"`       // request address on Binance chain
	Destination   string         `json:"destination"`     // destination , not sure what we use it for
	Amount        string         `json:"amount"`          // amount of token to swap
	Owner         sdk.AccAddress `json:"owner"`
}

// NewMsgSwap is a constructor function for MsgSwap
func NewMsgSwap(requestTxHash, source, target, amount, requester, destination string, owner sdk.AccAddress) MsgSwap {
	return MsgSwap{
		RequestTxHash: requestTxHash,
		SourceTicker:  source,
		TargetTicker:  target,
		Amount:        amount,
		Requester:     requester,
		Destination:   destination,
		Owner:         owner,
	}
}

// Route should return the pooldata of the module
func (msg MsgSwap) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSwap) Type() string { return "set_swap" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSwap) ValidateBasic() sdk.Error {
	if len(msg.RequestTxHash) == 0 {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}
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
