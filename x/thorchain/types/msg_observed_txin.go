package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgObservedTxIn defines a MsgObservedTxIn message
type MsgObservedTxIn struct {
	Txs    ObservedTxs    `json:"txs"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgObservedTxIn is a constructor function for MsgObservedTxIn
func NewMsgObservedTxIn(txs ObservedTxs, signer sdk.AccAddress) MsgObservedTxIn {
	return MsgObservedTxIn{
		Txs:    txs,
		Signer: signer,
	}
}

// Route should return the cmname of the module
func (msg MsgObservedTxIn) Route() string { return RouterKey }

// Type should return the action
func (msg MsgObservedTxIn) Type() string { return "set_observed_txin" }

// ValidateBasic runs stateless checks on the message
func (msg MsgObservedTxIn) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if len(msg.Txs) == 0 {
		return sdk.ErrUnknownRequest("Txs cannot be empty")
	}
	for _, tx := range msg.Txs {
		if err := tx.Valid(); err != nil {
			return sdk.ErrUnknownRequest(err.Error())
		}
		obAddr, err := tx.ObservedPubKey.GetAddress(tx.Tx.Coins[0].Asset.Chain)
		if err != nil {
			return sdk.ErrUnknownRequest(err.Error())
		}
		if !tx.Tx.ToAddress.Equals(obAddr) {
			return sdk.ErrUnknownRequest("Request is not an inbound observed transaction")
		}
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgObservedTxIn) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgObservedTxIn) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
