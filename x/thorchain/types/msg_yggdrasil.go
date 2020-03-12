package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// MsgYggdrasil defines a MsgYggdrasil message
type MsgYggdrasil struct {
	Tx          common.Tx      `json:"tx"`
	PubKey      common.PubKey  `json:"pub_key"`
	AddFunds    bool           `json:"add_funds"`
	Coins       common.Coins   `json:"coins"`
	BlockHeight int64          `json:"block_height"`
	Signer      sdk.AccAddress `json:"signer"`
}

// NewMsgYggdrasil is a constructor function for MsgYggdrasil
func NewMsgYggdrasil(tx common.Tx, pk common.PubKey, blockHeight int64, addFunds bool, coins common.Coins, signer sdk.AccAddress) MsgYggdrasil {
	return MsgYggdrasil{
		Tx:          tx,
		PubKey:      pk,
		AddFunds:    addFunds,
		Coins:       coins,
		BlockHeight: blockHeight,
		Signer:      signer,
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
	if msg.BlockHeight <= 0 {
		return sdk.ErrUnknownRequest("invalid block height")
	}
	if msg.Tx.IsEmpty() {
		return sdk.ErrUnknownRequest("request tx cannot be empty")
	}
	for _, coin := range msg.Coins {
		if err := coin.IsValid(); err != nil {
			return sdk.ErrInvalidCoins(err.Error())
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
