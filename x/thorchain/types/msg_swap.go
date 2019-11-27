package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

// MsgSwap defines a MsgSwap message
type MsgSwap struct {
	Tx          common.Tx      `json:"tx"`           // request tx
	TargetAsset common.Asset   `json:"target_asset"` // target asset
	Destination common.Address `json:"destination"`  // destination , used for swap and send , the destination address we send it to
	TradeTarget sdk.Uint       `json:"trade_target"`
	Signer      sdk.AccAddress `json:"signer"`
}

// NewMsgSwap is a constructor function for MsgSwap
func NewMsgSwap(tx common.Tx, target common.Asset, destination common.Address, tradeTarget sdk.Uint, signer sdk.AccAddress) MsgSwap {
	return MsgSwap{
		Tx:          tx,
		TargetAsset: target,
		Destination: destination,
		TradeTarget: tradeTarget,
		Signer:      signer,
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
	if err := msg.Tx.IsValid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	if msg.TargetAsset.IsEmpty() {
		return sdk.ErrUnknownRequest("Swap Target cannot be empty")
	}
	for _, coin := range msg.Tx.Coins {
		if coin.Asset.Equals(msg.TargetAsset) {
			return sdk.ErrUnknownRequest("Swap Source and Target cannot be the same.")
		}
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
