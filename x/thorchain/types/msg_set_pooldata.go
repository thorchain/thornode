package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

// MsgSetPoolData defines a SetPoolData message
// We keep this for now , as a mechanism to set up a new pool when it is not in the genesis file
// the pool changes when stake / swap happens
type MsgSetPoolData struct {
	BalanceRune  sdk.Uint       `json:"balance_rune"`  // balance rune
	BalanceAsset sdk.Uint       `json:"balance_asset"` // balance of asset
	Asset        common.Asset   `json:"asset"`         // Asset means the asset asset
	Status       PoolStatus     `json:"status"`        // pool status
	Signer       sdk.AccAddress `json:"signer"`
}

// NewMsgSetPoolData is a constructor function for MsgSetPoolData
func NewMsgSetPoolData(asset common.Asset, status PoolStatus, signer sdk.AccAddress) MsgSetPoolData {
	return MsgSetPoolData{
		Asset:        asset,
		BalanceRune:  sdk.ZeroUint(),
		BalanceAsset: sdk.ZeroUint(),
		Status:       status,
		Signer:       signer,
	}
}

// Route should return the pooldata of the module
func (msg MsgSetPoolData) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetPoolData) Type() string { return "set_pooldata" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetPoolData) ValidateBasic() sdk.Error {
	if msg.Asset.IsEmpty() {
		return sdk.ErrUnknownRequest("pool Asset cannot be empty")
	}
	if msg.Asset.IsRune() {
		return sdk.ErrUnknownRequest("invalid pool asset")
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
	return []sdk.AccAddress{msg.Signer}
}
