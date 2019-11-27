package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// MsgNextPoolAddress is used to set the pool address of the next
type MsgNextPoolAddress struct {
	Tx             common.Tx
	NextPoolPubKey common.PubKey
	Sender         common.Address
	Signer         sdk.AccAddress
	Chain          common.Chain
}

// NewMsgNextPoolAddress create a new instance of MsgNextPoolAddress
func NewMsgNextPoolAddress(tx common.Tx, nextPoolPubKey common.PubKey, sender common.Address, chain common.Chain, signer sdk.AccAddress) MsgNextPoolAddress {
	return MsgNextPoolAddress{
		Tx:             tx,
		NextPoolPubKey: nextPoolPubKey,
		Sender:         sender,
		Signer:         signer,
		Chain:          chain,
	}
}

// Route should return the router key of the module
func (msg MsgNextPoolAddress) Route() string { return RouterKey }

// Type should return the action
func (msg MsgNextPoolAddress) Type() string { return "set_next_pooladdress" }

// ValidateBasic runs stateless checks on the message
func (msg MsgNextPoolAddress) ValidateBasic() sdk.Error {
	if msg.NextPoolPubKey.IsEmpty() {
		return sdk.ErrUnknownRequest("next pool pub key cannot be empty")
	}
	if msg.Sender.IsEmpty() {
		return sdk.ErrUnknownRequest("sender cannot be empty")
	}
	if err := msg.Tx.IsValid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	if msg.Chain.IsEmpty() {
		return sdk.ErrUnknownRequest("chain cannot be empty")
	}
	if !common.IsBNBChain(msg.Chain) {
		return sdk.ErrUnknownRequest("nextpool memo will only happen on BNB chain")
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
