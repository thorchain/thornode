package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// MsgSetTxHashComplete defines a MsgSetTxHashComplete message
type MsgSetTxHashComplete struct {
	RequestTxHash  string         `json:"request_tx_hash"`
	CompleteTxHash string         `json:"complete_tx_hash"`
	Signer         sdk.AccAddress `json:"signer"`
}

// NewMsgSetTxHashComplete is a constructor function for MsgSetTxHashComplete
func NewMsgSetTxHashComplete(requestTxHash, completeTxHash string, signer sdk.AccAddress) MsgSetTxHashComplete {
	return MsgSetTxHashComplete{
		RequestTxHash:  requestTxHash,
		CompleteTxHash: completeTxHash,
		Signer:         signer,
	}
}

// Route should return the cmname of the module
func (msg MsgSetTxHashComplete) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetTxHashComplete) Type() string { return "set_tx_hash_complete" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetTxHashComplete) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if len(msg.RequestTxHash) == 0 {
		return sdk.ErrUnknownRequest("request hash cannot be empty")
	}
	if len(msg.CompleteTxHash) == 0 {
		return sdk.ErrUnknownRequest("complete tx hash cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetTxHashComplete) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetTxHashComplete) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
