package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgMigrate defines a MsgMigrate message
type MsgMigrate struct {
	Tx          ObservedTx     `json:"tx"`
	BlockHeight int64          `json:"block_height"`
	Signer      sdk.AccAddress `json:"signer"`
}

// NewMsgMigrate is a constructor function for MsgMigrate
func NewMsgMigrate(tx ObservedTx, blockHeight int64, signer sdk.AccAddress) MsgMigrate {
	return MsgMigrate{
		Tx:          tx,
		BlockHeight: blockHeight,
		Signer:      signer,
	}
}

// Route should return the name of the module
func (msg MsgMigrate) Route() string { return RouterKey }

// Type should return the action
func (msg MsgMigrate) Type() string { return "migrate" }

// ValidateBasic runs stateless checks on the message
func (msg MsgMigrate) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if msg.BlockHeight <= 0 {
		return sdk.ErrUnknownRequest("invalid block height")
	}
	if err := msg.Tx.Valid(); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgMigrate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgMigrate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
