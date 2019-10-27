package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

// MsgSetStakeData defines a SetStakeData message
type MsgSetStakeData struct {
	Asset         common.Asset   `json:"asset"`           // ticker means the asset
	AssetAmount   sdk.Uint       `json:"asset_amt"`       // the amount of asset stake
	RuneAmount    sdk.Uint       `json:"rune"`            // the amount of rune stake
	PublicAddress common.Address `json:"public_address"`  // Staker's address on chain
	RequestTxHash common.TxID    `json:"request_tx_hash"` // the txhash that represent user send asset to our pool address
	Signer        sdk.AccAddress `json:"signer"`
}

// NewMsgSetStakeData is a constructor function for MsgSetStakeData
func NewMsgSetStakeData(asset common.Asset, r, amount sdk.Uint, publicAddress common.Address, requestTxHash common.TxID, signer sdk.AccAddress) MsgSetStakeData {
	return MsgSetStakeData{
		Asset:         asset,
		AssetAmount:   amount,
		RuneAmount:    r,
		PublicAddress: publicAddress,
		RequestTxHash: requestTxHash,
		Signer:        signer,
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

	if msg.RequestTxHash.IsEmpty() {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}
	if msg.PublicAddress.IsEmpty() {
		return sdk.ErrUnknownRequest("public address cannot be empty")
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
