package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

// MsgLeave when an operator don't want to be a validator anymore
type MsgLeave struct {
	Tx     common.Tx      `json:"tx"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgLeave create a new instance of MsgLeave
func NewMsgLeave(tx common.Tx, signer sdk.AccAddress) MsgLeave {
	return MsgLeave{
		Tx:     tx,
		Signer: signer,
	}
}

// Route should return the router key of the module
func (msg MsgLeave) Route() string { return RouterKey }

// Type should return the action
func (msg MsgLeave) Type() string { return "validator_leave" }

// ValidateBasic runs stateless checks on the message
func (msg MsgLeave) ValidateBasic() sdk.Error {
	if msg.Tx.FromAddress.IsEmpty() {
		return sdk.ErrUnknownRequest("from address cannot be empty")
	}
	if msg.Tx.ID.IsEmpty() {
		return sdk.ErrUnknownRequest("tx id hash cannot be empty")
	}
	if msg.Signer.Empty() {
		return sdk.ErrUnknownRequest("signer cannot be empty ")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgLeave) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgLeave) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
