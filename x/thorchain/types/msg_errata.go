package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// MsgErrataTx defines a MsgErrataTx message
type MsgErrataTx struct {
	TxID   common.TxID    `json:"txid"`
	Chain  common.Chain   `json:"chain"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgErrataTx is a constructor function for NewMsgErrataTx
func NewMsgErrataTx(txID common.TxID, chain common.Chain, signer sdk.AccAddress) MsgErrataTx {
	return MsgErrataTx{
		TxID:   txID,
		Chain:  chain,
		Signer: signer,
	}
}

// Route should return the cmname of the module
func (msg MsgErrataTx) Route() string { return RouterKey }

// Type should return the action
func (msg MsgErrataTx) Type() string { return "errata_tx" }

// ValidateBasic runs stateless checks on the message
func (msg MsgErrataTx) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if msg.TxID.IsEmpty() {
		return sdk.ErrUnknownRequest("tx id cannot be empty")
	}
	if msg.Chain.IsEmpty() {
		return sdk.ErrUnknownRequest("chain cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgErrataTx) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgErrataTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
