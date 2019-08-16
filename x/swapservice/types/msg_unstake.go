package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MaxWithdrawBasisPoints
const MaxWithdrawBasisPoints = 10000

// MsgSetUnStake is used to withdraw
type MsgSetUnStake struct {
	PublicAddress       BnbAddress     `json:"public_address"`        // it should be the public address
	WithdrawBasisPoints Amount         `json:"withdraw_basis_points"` // withdraw basis points
	Ticker              Ticker         `json:"ticker"`                // ticker token symbol
	RequestTxHash       TxID           `json:"request_tx_hash"`       // request tx hash on binance chain
	Owner               sdk.AccAddress `json:"owner"`
}

// NewMsgSetUnStake is a constructor function for MsgSetPoolData
func NewMsgSetUnStake(publicAddress BnbAddress, withdrawBasisPoints Amount, ticker Ticker, requestTxHash TxID, owner sdk.AccAddress) MsgSetUnStake {
	return MsgSetUnStake{
		PublicAddress:       publicAddress,
		WithdrawBasisPoints: withdrawBasisPoints,
		Ticker:              ticker,
		RequestTxHash:       requestTxHash,
		Owner:               owner,
	}
}

// Route should return the pooldata of the module
func (msg MsgSetUnStake) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetUnStake) Type() string { return "set_unstake" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetUnStake) ValidateBasic() sdk.Error {
	if msg.Owner.Empty() {
		return sdk.ErrUnknownRequest("Owner cannot be empty")
	}
	if msg.Ticker.Empty() {
		return sdk.ErrUnknownRequest("Pool Ticker cannot be empty")
	}
	if msg.PublicAddress.Empty() {
		return sdk.ErrUnknownRequest("Address cannot be empty")
	}
	if msg.RequestTxHash.Empty() {
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
	return []sdk.AccAddress{msg.Owner}
}
