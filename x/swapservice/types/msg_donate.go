package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgDonate defines a donate message
type MsgDonate struct {
	Ticker      Ticker         `json:"ticker"` // ticker means the symbol
	TokenAmount Amount         `json:"token"`  // the amount of token
	RuneAmount  Amount         `json:"rune"`   // the amount of rune
	TxID        TxID           `json:"tx_id"`  // the txhash that represent user send token to our pool address
	Signer      sdk.AccAddress `json:"signer"`
}

// NewMsgDonate is a constructor function for MsgDonate
func NewMsgDonate(ticker Ticker, r, token Amount, requestTxHash TxID, signer sdk.AccAddress) MsgDonate {
	return MsgDonate{
		Ticker:      ticker,
		TokenAmount: token,
		RuneAmount:  r,
		TxID:        requestTxHash,
		Signer:      signer,
	}
}

// Route should return the pooldata of the module
func (msg MsgDonate) Route() string { return RouterKey }

// Type should return the action
func (msg MsgDonate) Type() string { return "set_donate" }

// ValidateBasic runs stateless checks on the message
func (msg MsgDonate) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrUnknownRequest("Signer cannot be empty")
	}
	if msg.Ticker.Empty() {
		return sdk.ErrUnknownRequest("Stake Ticker cannot be empty")
	}
	if msg.RuneAmount.Empty() {
		return sdk.ErrUnknownRequest("Stake Rune cannot be empty")
	}
	if msg.TokenAmount.Empty() {
		return sdk.ErrUnknownRequest("Stake Token cannot be empty")
	}
	if msg.TxID.Empty() {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgDonate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgDonate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
