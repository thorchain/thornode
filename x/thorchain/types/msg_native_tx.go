package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	common "gitlab.com/thorchain/thornode/common"
)

// MsgNativeTx defines a MsgNativeTx message
type MsgNativeTx struct {
	Coins  common.Coins   `json:"coins"`
	Memo   string         `json:"memo"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgNativeTx is a constructor function for NewMsgNativeTx
func NewMsgNativeTx(coins common.Coins, memo string, signer sdk.AccAddress) MsgNativeTx {
	return MsgNativeTx{
		Coins:  coins,
		Memo:   memo,
		Signer: signer,
	}
}

// Route should return the cmname of the module
func (msg MsgNativeTx) Route() string { return RouterKey }

// Type should return the action
func (msg MsgNativeTx) Type() string { return "native_tx" }

// ValidateBasic runs stateless checks on the message
func (msg MsgNativeTx) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if err := msg.Coins.IsValid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	for _, coin := range msg.Coins {
		if !coin.IsNative() {
			return sdk.ErrUnknownRequest("all coins must be native to THORChain")
		}
	}
	if len([]byte(msg.Memo)) > 150 {
		err := fmt.Errorf("Memo must not exceed 150 bytes: %d", len([]byte(msg.Memo)))
		return sdk.ErrUnknownRequest(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgNativeTx) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgNativeTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
