package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// MsgBond when a user would like to become a validator, and run a full set, they need send an `apply:bepaddress` with a bond to our pool address
type MsgBond struct {
	TxIn        common.Tx      `json:"tx_in"`
	NodeAddress sdk.AccAddress `json:"node_address"`
	Bond        sdk.Uint       `json:"bond"`
	BondAddress common.Address `json:"bond_address"`
	Signer      sdk.AccAddress `json:"signer"`
}

// NewMsgBond create new MsgBond message
func NewMsgBond(txin common.Tx, nodeAddr sdk.AccAddress, bond sdk.Uint, bondAddress common.Address, signer sdk.AccAddress) MsgBond {
	return MsgBond{
		TxIn:        txin,
		NodeAddress: nodeAddr,
		Bond:        bond,
		BondAddress: bondAddress,
		Signer:      signer,
	}
}

// Route should return the router key of the module
func (msg MsgBond) Route() string { return RouterKey }

// Type should return the action
func (msg MsgBond) Type() string { return "validator_apply" }

// ValidateBasic runs stateless checks on the message
func (msg MsgBond) ValidateBasic() sdk.Error {
	if msg.NodeAddress.Empty() {
		return sdk.ErrUnknownRequest("node address cannot be empty")
	}
	if msg.Bond.IsZero() {
		return sdk.ErrUnknownRequest("bond cannot be zero")
	}
	if msg.BondAddress.IsEmpty() {
		return sdk.ErrUnknownRequest("bond address cannot be empty")
	}
	if msg.TxIn.IsEmpty() {
		return sdk.ErrUnknownRequest("request tx cannot be empty")
	}
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress("empty signer address")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgBond) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgBond) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
