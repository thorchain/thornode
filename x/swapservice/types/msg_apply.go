package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
)

// MsgApply when a user would like to become a validator, and run a full set, they need send an `apply:bepaddress` with a bond to our pool address
type MsgApply struct {
	NodeAddress   sdk.AccAddress `json:"node_address"`
	Bond          sdk.Uint       `json:"bond"`
	RequestTxHash common.TxID    `json:"request_tx_hash"` // request tx hash on binance chain
	Signer        sdk.AccAddress `json:"signer"`
}

// NewMsgApply create new MsgApply message
func NewMsgApply(nodeAddr sdk.AccAddress, bond sdk.Uint, requestTxHash common.TxID, signer sdk.AccAddress) MsgApply {
	return MsgApply{
		NodeAddress:   nodeAddr,
		Bond:          bond,
		RequestTxHash: requestTxHash,
		Signer:        signer,
	}
}

// Route should return the router key of the module
func (msg MsgApply) Route() string { return RouterKey }

// Type should return the action
func (msg MsgApply) Type() string { return "validator_apply" }

// ValidateBasic runs stateless checks on the message
func (msg MsgApply) ValidateBasic() sdk.Error {
	if msg.NodeAddress.Empty() {
		return sdk.ErrUnknownRequest("node address cannot be empty")
	}
	if msg.Bond.IsZero() {
		return sdk.ErrUnknownRequest("bond cannot be zero")
	}
	if msg.RequestTxHash.IsEmpty() {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}
	if msg.Signer.Empty() {
		return sdk.ErrUnknownRequest("signer cannot be empty ")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgApply) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgApply) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
