package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// MsgSetTxIn defines a MsgSetTxIn message
type MsgSetTxIn struct {
	TxIns  []TxIn         `json:"tx_hashes"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgSetTxIn is a constructor function for MsgSetTxIn
func NewMsgSetTxIn(txs []TxIn, signer sdk.AccAddress) MsgSetTxIn {
	return MsgSetTxIn{
		TxIns:  txs,
		Signer: signer,
	}
}

// Route should return the cmname of the module
func (msg MsgSetTxIn) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetTxIn) Type() string { return "set_tx_hashes" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetTxIn) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if len(msg.TxIns) == 0 {
		return sdk.ErrUnknownRequest("Tx Hashes cannot be empty")
	}
	for _, tx := range msg.TxIns {
		if err := tx.Valid(); err != nil {
			return sdk.ErrUnknownRequest(err.Error())
		}
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetTxIn) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetTxIn) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
