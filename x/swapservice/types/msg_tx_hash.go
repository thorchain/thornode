package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// MsgSetTxHash defines a MsgSetTxHash message
type MsgSetTxHash struct {
	TxHash TxHash         `json:"tx_hash"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgSetTxHash is a constructor function for MsgSetTxHash
func NewMsgSetTxHash(txHash string, signer sdk.AccAddress) MsgSetTxHash {
	return MsgSetTxHash{
		TxHash: NewTxHash(txHash),
		Signer: signer,
	}
}

// Route should return the cmname of the module
func (msg MsgSetTxHash) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetTxHash) Type() string { return "set_tx_hash" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetTxHash) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if len(msg.TxHash.TxHash) == 0 {
		return sdk.ErrUnknownRequest("Tx hash cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetTxHash) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetTxHash) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
