package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgUnStakeComplete set an unstake to complete state
type MsgUnStakeComplete struct {
	RequestTxHash  string `json:"request_tx_hash"`  // the request tx hash from binance chain to request the unstake
	CompleteTxHash string `json:"complete_tx_hash"` // the tx hash we send token back to user on binance chain
	Owner          sdk.AccAddress
}

// NewMsgUnStakeComplete create a new instance of MsgUnStakeComplete
// the message is used to mark a unstake as complete record the tx hash on binance chain which indicate we pay user accordingly for audit purpose
func NewMsgUnStakeComplete(requestTxHash, completeTxHash string, owner sdk.AccAddress) MsgUnStakeComplete {
	return MsgUnStakeComplete{
		RequestTxHash:  requestTxHash,
		CompleteTxHash: completeTxHash,
		Owner:          owner,
	}
}

// Route should return the pooldata of the module
func (msg MsgUnStakeComplete) Route() string { return RouterKey }

// Type should return the action
func (msg MsgUnStakeComplete) Type() string { return "set_unstake_complete" }

// ValidateBasic runs stateless checks on the message
func (msg MsgUnStakeComplete) ValidateBasic() sdk.Error {
	if len(msg.RequestTxHash) == 0 {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}
	if len(msg.CompleteTxHash) == 0 {
		return sdk.ErrUnknownRequest("tx hash for paying user can't be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgUnStakeComplete) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgUnStakeComplete) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}
