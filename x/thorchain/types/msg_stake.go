package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

// MsgSetStakeData defines a SetStakeData message
type MsgSetStakeData struct {
	Tx           common.Tx      `json:"tx"`
	Asset        common.Asset   `json:"asset"`         // ticker means the asset
	AssetAmount  sdk.Uint       `json:"asset_amt"`     // the amount of asset stake
	RuneAmount   sdk.Uint       `json:"rune"`          // the amount of rune stake
	RuneAddress  common.Address `json:"rune_address"`  // staker's rune address
	AssetAddress common.Address `json:"asset_address"` // staker's asset address
	Signer       sdk.AccAddress `json:"signer"`
}

// NewMsgSetStakeData is a constructor function for MsgSetStakeData
func NewMsgSetStakeData(tx common.Tx, asset common.Asset, r, amount sdk.Uint, runeAddr, assetAddr common.Address, signer sdk.AccAddress) MsgSetStakeData {
	return MsgSetStakeData{
		Tx:           tx,
		Asset:        asset,
		AssetAmount:  amount,
		RuneAmount:   r,
		RuneAddress:  runeAddr,
		AssetAddress: assetAddr,
		Signer:       signer,
	}
}

// Route should return the pooldata of the module
func (msg MsgSetStakeData) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetStakeData) Type() string { return "set_stakedata" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetStakeData) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if msg.Asset.IsEmpty() {
		return sdk.ErrUnknownRequest("Stake asset cannot be empty")
	}
	if err := msg.Tx.IsValid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	if msg.RuneAddress.IsEmpty() {
		return sdk.ErrUnknownRequest("rune address cannot be empty")
	}
	if !msg.Asset.Chain.IsBNB() {
		if msg.AssetAddress.IsEmpty() {
			return sdk.ErrUnknownRequest("asset address cannot be empty")
		}

	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetStakeData) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetStakeData) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
