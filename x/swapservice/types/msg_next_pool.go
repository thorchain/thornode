package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
)

// MsgNextPoolAddress is used to set the pool address of the next
type MsgNextPoolAddress struct {
	RequestTxHash common.TxID
	Sender        common.BnbAddress
	Signer        sdk.AccAddress
}

// NewMsgNextPoolAddress create a new instance of MsgNextPoolAddress
func NewMsgNextPoolAddress(requestTxHash common.TxID, sender common.BnbAddress, signer sdk.AccAddress) MsgNextPoolAddress {
	return MsgNextPoolAddress{
		RequestTxHash: requestTxHash,
		Sender:        sender,
		Signer:        signer,
	}
}

// Route should return the router key of the module
func (msg MsgNextPoolAddress) Route() string { return RouterKey }

// Type should return the action
func (msg MsgNextPoolAddress) Type() string { return "set_next_pooladdress" }

// ValidateBasic runs stateless checks on the message
func (msg MsgNextPoolAddress) ValidateBasic() sdk.Error {
	if msg.Sender.IsEmpty() {
		return sdk.ErrUnknownRequest("sender can't be empty")
	}
	if msg.RequestTxHash.IsEmpty() {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgNextPoolAddress) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgNextPoolAddress) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
