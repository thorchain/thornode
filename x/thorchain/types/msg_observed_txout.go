package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgObservedTxOut defines a MsgObservedTxOut message
type MsgObservedTxOut struct {
	Txs    ObservedTxs    `json:"txs"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgObservedTxOut is a constructor function for MsgObservedTxOut
func NewMsgObservedTxOut(txs ObservedTxs, signer sdk.AccAddress) MsgObservedTxOut {
	return MsgObservedTxOut{
		Txs:    txs,
		Signer: signer,
	}
}

// Route should return the cmname of the module
func (msg MsgObservedTxOut) Route() string { return RouterKey }

// Type should return the action
func (msg MsgObservedTxOut) Type() string { return "set_observed_txout" }

// ValidateBasic runs stateless checks on the message
func (msg MsgObservedTxOut) ValidateBasic() sdk.Error {
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
		if !tx.Tx.FromAddress.Equals(obAddr) {
			return sdk.ErrUnknownRequest("Request is not an outbound observed transaction")
		}
		if len(tx.Signers) > 0 {
			return sdk.ErrUnknownRequest("signers must be empty")
		}
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgObservedTxOut) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgObservedTxOut) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
