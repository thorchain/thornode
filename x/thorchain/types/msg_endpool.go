package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

type MsgEndPool struct {
	Asset         common.Asset   `json:"asset"`
	Requester     common.Address `json:"requester"`
	RequestTxHash common.TxID    `json:"request_tx_hash"` // request tx hash on chain
	Signer        sdk.AccAddress `json:"signer"`
}

// NewMsgEndPool create a new instance MsgEndPool
func NewMsgEndPool(asset common.Asset, request common.Address, requestTxHash common.TxID, signer sdk.AccAddress) MsgEndPool {
	return MsgEndPool{
		Asset:         asset,
		Requester:     request,
		RequestTxHash: requestTxHash,
		Signer:        signer,
	}
}

// Route should return the router key of the module
func (msg MsgEndPool) Route() string { return RouterKey }

// Type should return the action
func (msg MsgEndPool) Type() string { return "set_poolend" }

// ValidateBasic runs stateless checks on the message
func (msg MsgEndPool) ValidateBasic() sdk.Error {
	if msg.Asset.IsEmpty() {
		return sdk.ErrUnknownRequest("pool Asset cannot be empty")
	}
	if msg.Asset.IsRune() {
		return sdk.ErrUnknownRequest("invalid pool asset")
	}
	if msg.RequestTxHash.IsEmpty() {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}
	if msg.Requester.IsEmpty() {
		return sdk.ErrUnknownRequest("invalid requester")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgEndPool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgEndPool) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
