package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

// MsgYggdrasil defines a MsgYggdrasil message
type MsgYggdrasil struct {
	PubKey        common.PubKey  `json:"pub_key"`
	AddFunds      bool           `json:"add_funds"`
	Coins         common.Coins   `json:"coins"`
	RequestTxHash common.TxID    `json:"request_tx_hash"`
	Signer        sdk.AccAddress `json:"signer"`
}

// NewMsgYggdrasil is a constructor function for MsgYggdrasil
func NewMsgYggdrasil(pk common.PubKey, addFunds bool, coins common.Coins, requestTxHash common.TxID, signer sdk.AccAddress) MsgYggdrasil {
	return MsgYggdrasil{
		PubKey:        pk,
		AddFunds:      addFunds,
		Coins:         coins,
		RequestTxHash: requestTxHash,
		Signer:        signer,
	}
}

func (msg MsgYggdrasil) Route() string { return RouterKey }

func (msg MsgYggdrasil) Type() string { return "set_yggdrasil" }

// ValidateBasic runs stateless checks on the message
func (msg MsgYggdrasil) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if msg.PubKey.IsEmpty() {
		return sdk.ErrUnknownRequest("pubkey cannot be empty")
	}
	if msg.RequestTxHash.IsEmpty() {
		return sdk.ErrUnknownRequest("request tx hash cannot be empty")
	}
	for _, coin := range msg.Coins {
		if err := coin.IsValid(); err != nil {
			return sdk.ErrUnknownRequest(err.Error())
		}
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgYggdrasil) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgYggdrasil) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
