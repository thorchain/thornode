package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSwapComplete set the a swap to complete state
type MsgSwapComplete struct {
	RequestTxHash string // the request tx hash from binance chain
	PayTxHash     string // the tx hash indicate we pay to user's account
	Owner         sdk.AccAddress
}

// NewMsgSwapComplete create a new instance of MsgSwapComplete
// The message we use to mark a swap as complete, record the tx hash on binance chain
// which indicate we pay user accordingly , for audit purpose later.
func NewMsgSwapComplete(requestTxHash, payTxHash string, owner sdk.AccAddress) MsgSwapComplete {
	return MsgSwapComplete{
		RequestTxHash: requestTxHash,
		PayTxHash:     payTxHash,
		Owner:         owner,
	}
}

// Route should return the pooldata of the module
func (msg MsgSwapComplete) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSwapComplete) Type() string { return "set_swap_complete" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSwapComplete) ValidateBasic() sdk.Error {
	if len(msg.RequestTxHash) == 0 {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}
	if len(msg.PayTxHash) == 0 {
		return sdk.ErrUnknownRequest("tx hash for paying user can't be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSwapComplete) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSwapComplete) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}
