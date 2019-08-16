package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	common "gitlab.com/thorchain/bepswap/common"
)

// MsgSetStakeData defines a SetStakeData message
type MsgSetStakeData struct {
	Ticker        common.Ticker     `json:"ticker"`          // ticker means the symbol
	TokenAmount   common.Amount     `json:"token"`           // the amount of token stake
	RuneAmount    common.Amount     `json:"rune"`            // the amount of rune stake
	PublicAddress common.BnbAddress `json:"public_address"`  // Staker's address on binance chain
	RequestTxHash common.TxID       `json:"request_tx_hash"` // the txhash that represent user send token to our pool address
	Signer        sdk.AccAddress    `json:"signer"`
}

// NewMsgSetStakeData is a constructor function for MsgSetStakeData
func NewMsgSetStakeData(ticker common.Ticker, r, token common.Amount, publicAddress common.BnbAddress, requestTxHash common.TxID, signer sdk.AccAddress) MsgSetStakeData {
	return MsgSetStakeData{
		Ticker:        ticker,
		TokenAmount:   token,
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
	if msg.Ticker.IsEmpty() {
		return sdk.ErrUnknownRequest("Stake Ticker cannot be empty")
	}
	if msg.RuneAmount.IsEmpty() {
		return sdk.ErrUnknownRequest("Stake Rune cannot be empty")
	}
	if msg.TokenAmount.IsEmpty() {
		return sdk.ErrUnknownRequest("Stake Token cannot be empty")
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
