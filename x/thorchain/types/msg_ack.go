package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// MsgAck is used to confirm the next pool address
type MsgAck struct {
	Tx     common.Tx
	Sender common.Address
	Signer sdk.AccAddress
	Chain  common.Chain // which chain this ack is from
}

// NewMsgAck create a new instance of NewMsgAck
func NewMsgAck(tx common.Tx, sender common.Address, chain common.Chain, signer sdk.AccAddress) MsgAck {
	return MsgAck{
		Tx:     tx,
		Sender: sender,
		Signer: signer,
		Chain:  chain,
	}
}

// Route should return the router key of the module
func (msg MsgAck) Route() string { return RouterKey }

// Type should return the action
func (msg MsgAck) Type() string { return "set_ack" }

// ValidateBasic runs stateless checks on the message
func (msg MsgAck) ValidateBasic() sdk.Error {
	if msg.Sender.IsEmpty() {
		return sdk.ErrUnknownRequest("sender address cannot be empty")
	}
	if err := msg.Tx.IsValid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	if msg.Chain.IsEmpty() {
		return sdk.ErrUnknownRequest("chain cannot be empty")
	}

	if !msg.Chain.IsBNB() {
		return sdk.ErrUnknownRequest("ack memo only happen on BNB chan")
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgAck) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgAck) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
