package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const RouterKey = ModuleName // this was defined in your key.go file

// MsgSetPool defines a MsgSetPool message
type MsgSetPool struct {
	Address     sdk.AccAddress `json:"address"`
	TokenName   string         `json:"token_name"`
	TokenTicker string         `json:"token_ticker"`
	Signer      sdk.AccAddress `json:"signer"`
}

// NewMsgSetPool is a constructor function for MsgSetPool
func NewMsgSetPool(name, ticker string, address, signer sdk.AccAddress) MsgSetPool {
	return MsgSetPool{
		Address:     address,
		TokenName:   name,
		TokenTicker: strings.ToUpper(ticker),
		Signer:      signer,
	}
}

// Route should return the cmname of the module
func (msg MsgSetPool) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetPool) Type() string { return "set_pool" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetPool) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if msg.Address.Empty() {
		return sdk.ErrInvalidAddress(msg.Address.String())
	}
	if len(msg.TokenName) == 0 {
		return sdk.ErrUnknownRequest("Token name cannot be empty")
	}
	if len(msg.TokenTicker) == 0 {
		return sdk.ErrUnknownRequest("Token ticker cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetPool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetPool) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// MsgSetTxHash defines a MsgSetTxHash message
type MsgSetTxHash struct {
	TxHash string         `json:"tx_hash"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgSetTxHash is a constructor function for MsgSetTxHash
func NewMsgSetTxHash(txHash string, address, signer sdk.AccAddress) MsgSetTxHash {
	return MsgSetTxHash{
		TxHash: txHash,
		Signer: signer,
	}
}

// Route should return the cmname of the module
func (msg MsgSetTxHash) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetTxHash) Type() string { return "set_tx_hash" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetTxHash) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if len(msg.TxHash) == 0 {
		return sdk.ErrUnknownRequest("Token ticker cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetTxHash) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetTxHash) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
