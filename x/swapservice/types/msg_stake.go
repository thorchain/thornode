package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSetStakeData defines a SetStakeData message
type MsgSetStakeData struct {
	Ticker        Ticker         `json:"ticker"`          // ticker means the symbol
	TokenAmount   Amount         `json:"token"`           // the amount of token stake
	RuneAmount    Amount         `json:"rune"`            // the amount of rune stake
	PublicAddress BnbAddress     `json:"public_address"`  // Staker's address on binance chain
	RequestTxHash TxID           `json:"request_tx_hash"` // the txhash that represent user send token to our pool address
	Owner         sdk.AccAddress `json:"owner"`
}

// NewMsgSetStakeData is a constructor function for MsgSetStakeData
func NewMsgSetStakeData(ticker Ticker, r, token Amount, publicAddress BnbAddress, requestTxHash TxID, owner sdk.AccAddress) MsgSetStakeData {
	return MsgSetStakeData{
		Ticker:        ticker,
		TokenAmount:   token,
		RuneAmount:    r,
		PublicAddress: publicAddress,
		RequestTxHash: requestTxHash,
		Owner:         owner,
	}
}

// Route should return the pooldata of the module
func (msg MsgSetStakeData) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetStakeData) Type() string { return "set_stakedata" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetStakeData) ValidateBasic() sdk.Error {
	if msg.Ticker.Empty() {
		return sdk.ErrUnknownRequest("Stake Ticker cannot be empty")
	}
	if msg.RuneAmount.Empty() {
		return sdk.ErrUnknownRequest("Stake Rune cannot be empty")
	}
	if msg.TokenAmount.Empty() {
		return sdk.ErrUnknownRequest("Stake Token cannot be empty")
	}
	if msg.RequestTxHash.Empty() {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}
	if msg.PublicAddress.Empty() {
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
	return []sdk.AccAddress{msg.Owner}
}
