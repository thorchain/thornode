package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

// MsgSwap defines a MsgSwap message
type MsgSwap struct {
	RequestTxHash common.TxID    `json:"request_tx_hash"` // Request transaction hash on chain
	SourceAsset   common.Asset   `json:"source_asset"`    // source asset
	TargetAsset   common.Asset   `json:"target_asset"`    // target asset
	Requester     common.Address `json:"requester"`       // request address on chain
	Destination   common.Address `json:"destination"`     // destination , used for swap and send , the destination address we send it to
	Amount        sdk.Uint       `json:"amount"`          // amount of asset to swap
	TradeTarget   sdk.Uint       `json:"trade_target"`
	Signer        sdk.AccAddress `json:"signer"`
}

// NewMsgSwap is a constructor function for MsgSwap
func NewMsgSwap(requestTxHash common.TxID, source, target common.Asset, amount sdk.Uint, requester, destination common.Address, tradeTarget sdk.Uint, signer sdk.AccAddress) MsgSwap {
	return MsgSwap{
		RequestTxHash: requestTxHash,
		SourceAsset:   source,
		TargetAsset:   target,
		Amount:        amount,
		Requester:     requester,
		Destination:   destination,
		TradeTarget:   tradeTarget,
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
	if msg.SourceAsset.IsEmpty() {
		return sdk.ErrUnknownRequest("Swap Source Asset cannot be empty")
	}
	if msg.TargetAsset.IsEmpty() {
		return sdk.ErrUnknownRequest("Swap Target cannot be empty")
	}
	if msg.SourceAsset.Equals(msg.TargetAsset) {
		return sdk.ErrUnknownRequest("Swap Source and Target cannot be the same.")
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
