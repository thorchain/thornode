package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// MaxUnstakeBasisPoints
const MaxUnstakeBasisPoints = 10_000

// MsgSetUnStake is used to withdraw
type MsgSetUnStake struct {
	Tx                 common.Tx      `json:"tx"`
	RuneAddress        common.Address `json:"rune_address"`          // it should be the rune address
	UnstakeBasisPoints sdk.Uint       `json:"withdraw_basis_points"` // withdraw basis points
	Asset              common.Asset   `json:"asset"`                 // asset asset asset
	Signer             sdk.AccAddress `json:"signer"`
}

// NewMsgSetUnStake is a constructor function for MsgSetPoolData
func NewMsgSetUnStake(tx common.Tx, runeAddress common.Address, withdrawBasisPoints sdk.Uint, asset common.Asset, signer sdk.AccAddress) MsgSetUnStake {
	return MsgSetUnStake{
		Tx:                 tx,
		RuneAddress:        runeAddress,
		UnstakeBasisPoints: withdrawBasisPoints,
		Asset:              asset,
		Signer:             signer,
	}
}

// Route should return the pooldata of the module
func (msg MsgSetUnStake) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetUnStake) Type() string { return "set_unstake" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetUnStake) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if err := msg.Tx.IsValid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	if msg.Asset.IsEmpty() {
		return sdk.ErrUnknownRequest("Pool Asset cannot be empty")
	}
	if msg.RuneAddress.IsEmpty() {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	if !msg.RuneAddress.IsChain(common.RuneAsset().Chain) {
		return sdk.ErrUnknownRequest(fmt.Sprintf("Address must be a %s address", common.RuneAsset().Chain))
	}
	if msg.UnstakeBasisPoints.IsZero() {
		return sdk.ErrUnknownRequest("UnstakeBasicPoints can't be zero")
	}
	if msg.UnstakeBasisPoints.GT(sdk.ZeroUint()) && msg.UnstakeBasisPoints.GT(sdk.NewUint(MaxUnstakeBasisPoints)) {
		return sdk.ErrUnknownRequest("UnstakeBasisPoints is larger than maximum withdraw basis points")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetUnStake) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetUnStake) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
