package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
)

type MsgEndPool struct {
	Ticker        common.Ticker  `json:"ticker"`
	Requester     common.Address `json:"requester"`
	RequestTxHash common.TxID    `json:"request_tx_hash"` // request tx hash on binance chain
	Signer        sdk.AccAddress `json:"signer"`
}

// NewMsgEndPool create a new instance MsgEndPool
func NewMsgEndPool(ticker common.Ticker, request common.Address, requestTxHash common.TxID, signer sdk.AccAddress) MsgEndPool {
	return MsgEndPool{
		Ticker:        ticker,
		Requester:     request,
		RequestTxHash: requestTxHash,
		Signer:        signer,
	}
}

// Route should return the router key of the module
func (msg MsgEndPool) Route() string { return RouterKey }

// Type should return the action
func (msg MsgEndPool) Type() string { return "set_poolend" }

// ValidateBasic runs stateless checks on the message
func (msg MsgEndPool) ValidateBasic() sdk.Error {
	if msg.Ticker.IsEmpty() {
		return sdk.ErrUnknownRequest("pool Ticker cannot be empty")
	}
	if common.IsRune(msg.Ticker) {
		return sdk.ErrUnknownRequest("invalid pool ticker")
	}
	if msg.RequestTxHash.IsEmpty() {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}
	if msg.Requester.IsEmpty() {
		return sdk.ErrUnknownRequest("invalid requester")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgEndPool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgEndPool) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
