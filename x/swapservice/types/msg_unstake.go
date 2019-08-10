package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSetUnStake is used to withdraw
type MsgSetUnStake struct {
	PublicAddress BnbAddress     `json:"public_address"`  // it should be the public address
	Percentage    Amount         `json:"percentage"`      // unstake percentage
	Ticker        Ticker         `json:"ticker"`          // ticker token symbol
	RequestTxHash TxID           `json:"request_tx_hash"` // request tx hash on binance chain
	Owner         sdk.AccAddress `json:"owner"`
}

// NewMsgSetUnStake is a constructor function for MsgSetPoolData
func NewMsgSetUnStake(publicAddress BnbAddress, percentage Amount, ticker Ticker, requestTxHash TxID, owner sdk.AccAddress) MsgSetUnStake {
	return MsgSetUnStake{
		PublicAddress: publicAddress,
		Percentage:    percentage,
		Ticker:        ticker,
		RequestTxHash: requestTxHash,
		Owner:         owner,
	}
}

// Route should return the pooldata of the module
func (msg MsgSetUnStake) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetUnStake) Type() string { return "set_unstake" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetUnStake) ValidateBasic() sdk.Error {
	if msg.Ticker.Empty() {
		return sdk.ErrUnknownRequest("Pool Ticker cannot be empty")
	}
	if msg.RequestTxHash.Empty() {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}
	if msg.Percentage.Empty() {
		return sdk.ErrUnknownRequest("Percentage cannot be empty")
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
