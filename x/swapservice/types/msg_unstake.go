package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	common "gitlab.com/thorchain/bepswap/common"
)

// MaxWithdrawBasisPoints
const MaxWithdrawBasisPoints = 10000

// MsgSetUnStake is used to withdraw
type MsgSetUnStake struct {
	PublicAddress       common.BnbAddress `json:"public_address"`        // it should be the public address
	WithdrawBasisPoints common.Amount     `json:"withdraw_basis_points"` // withdraw basis points
	Ticker              common.Ticker     `json:"ticker"`                // ticker token symbol
	RequestTxHash       common.TxID       `json:"request_tx_hash"`       // request tx hash on binance chain
	Signer              sdk.AccAddress    `json:"signer"`
}

// NewMsgSetUnStake is a constructor function for MsgSetPoolData
func NewMsgSetUnStake(publicAddress common.BnbAddress, withdrawBasisPoints common.Amount, ticker common.Ticker, requestTxHash common.TxID, signer sdk.AccAddress) MsgSetUnStake {
	return MsgSetUnStake{
		PublicAddress:       publicAddress,
		WithdrawBasisPoints: withdrawBasisPoints,
		Ticker:              ticker,
		RequestTxHash:       requestTxHash,
		Signer:              signer,
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
	if msg.Ticker.IsEmpty() {
		return sdk.ErrUnknownRequest("Pool Ticker cannot be empty")
	}
	if msg.PublicAddress.IsEmpty() {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	if msg.RequestTxHash.IsEmpty() {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}
	if msg.WithdrawBasisPoints.IsNegative() {
		return sdk.ErrUnknownRequest("withdraw basis points is invalid")
	}
	if msg.WithdrawBasisPoints.GreaterThen(0) && msg.WithdrawBasisPoints.Float64() > MaxWithdrawBasisPoints {
		return sdk.ErrUnknownRequest("WithdrawBasisPoints is larger than maximum withdraw basis points")
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
