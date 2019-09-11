package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
)

// MsgSwap defines a MsgSwap message
type MsgSwap struct {
	RequestTxHash common.TxID       `json:"request_tx_hash"` // Request transaction hash on Binance chain
	SourceTicker  common.Ticker     `json:"source_ticker"`   // source token
	TargetTicker  common.Ticker     `json:"target_ticker"`   // target token
	Requester     common.BnbAddress `json:"requester"`       // request address on Binance chain
	Destination   common.BnbAddress `json:"destination"`     // destination , used for swap and send , the destination address we send it to
	Amount        sdk.Uint          `json:"amount"`          // amount of token to swap
	TargetPrice   sdk.Uint          `json:"target_price"`
	Signer        sdk.AccAddress    `json:"signer"`
}

// NewMsgSwap is a constructor function for MsgSwap
func NewMsgSwap(requestTxHash common.TxID, source, target common.Ticker, amount sdk.Uint, requester, destination common.BnbAddress, targetPrice sdk.Uint, signer sdk.AccAddress) MsgSwap {
	return MsgSwap{
		RequestTxHash: requestTxHash,
		SourceTicker:  source,
		TargetTicker:  target,
		Amount:        amount,
		Requester:     requester,
		Destination:   destination,
		TargetPrice:   targetPrice,
		Signer:        signer,
	}
}

// Route should return the pooldata of the module
func (msg MsgSwap) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSwap) Type() string { return "set_swap" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSwap) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if msg.RequestTxHash.IsEmpty() {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}
	if msg.SourceTicker.IsEmpty() {
		return sdk.ErrUnknownRequest("Swap Source Ticker cannot be empty")
	}
	if msg.TargetTicker.IsEmpty() {
		return sdk.ErrUnknownRequest("Swap Target cannot be empty")
	}
	if msg.Amount.IsZero() {
		return sdk.ErrUnknownRequest("Swap Amount cannot be zero")
	}
	if msg.Requester.IsEmpty() {
		return sdk.ErrUnknownRequest("Swap Requester cannot be empty")
	}
	if msg.Destination.IsEmpty() {
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
	return []sdk.AccAddress{msg.Signer}
}
