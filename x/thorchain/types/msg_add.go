package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	common "gitlab.com/thorchain/bepswap/thornode/common"
)

// MsgAdd defines a add message
type MsgAdd struct {
	Asset       common.Asset   `json:"asset"`     // asset of the asset
	AssetAmount sdk.Uint       `json:"asset_amt"` // the amount of asset
	RuneAmount  sdk.Uint       `json:"rune"`      // the amount of rune
	TxID        common.TxID    `json:"tx_id"`     // the txhash that represent user send asset to our pool address
	Signer      sdk.AccAddress `json:"signer"`
}

// NewMsgAdd is a constructor function for MsgAdd
func NewMsgAdd(asset common.Asset, r, amount sdk.Uint, requestTxHash common.TxID, signer sdk.AccAddress) MsgAdd {
	return MsgAdd{
		Asset:       asset,
		AssetAmount: amount,
		RuneAmount:  r,
		TxID:        requestTxHash,
		Signer:      signer,
	}
}

// Route should return the pooldata of the module
func (msg MsgAdd) Route() string { return RouterKey }

// Type should return the action
func (msg MsgAdd) Type() string { return "set_add" }

// ValidateBasic runs stateless checks on the message
func (msg MsgAdd) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if msg.Asset.IsEmpty() {
		return sdk.ErrUnknownRequest("Add Asset cannot be empty")
	}
	if msg.TxID.IsEmpty() {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgAdd) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgAdd) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
