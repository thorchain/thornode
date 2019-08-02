package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const RouterKey = ModuleName // this was defined in your key.go file

// MsgSetPool defines a MsgSetPool message
type MsgSetPool struct {
	Pool   Pool           `json:"pool"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgSetPool is a constructor function for MsgSetPool
func NewMsgSetPool(name, ticker string, signer sdk.AccAddress) MsgSetPool {
	return MsgSetPool{
		Pool:   NewPool(name, ticker),
		Signer: signer,
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
	if msg.Pool.Address.Empty() {
		return sdk.ErrInvalidAddress(msg.Pool.Address.String())
	}
	if len(msg.Pool.TokenName) == 0 {
		return sdk.ErrUnknownRequest("Token name cannot be empty")
	}
	if len(msg.Pool.TokenTicker) == 0 {
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
	TxHash TxHash         `json:"tx_hash"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgSetTxHash is a constructor function for MsgSetTxHash
func NewMsgSetTxHash(txHash string, signer sdk.AccAddress) MsgSetTxHash {
	return MsgSetTxHash{
		TxHash: NewTxHash(txHash),
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
	if len(msg.TxHash.TxHash) == 0 {
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

// MsgSetUnStake defines a MsgSetUnStake message
type MsgSetUnStake struct {
	Coins  sdk.Coins      `json:"coins"`
	To     sdk.AccAddress `json:"to"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgSetUnStake is a constructor function for MsgSetUnStake
func NewMsgSetUnStake(coins sdk.Coins, to, signer sdk.AccAddress) MsgSetUnStake {
	return MsgSetUnStake{
		Coins:  coins,
		To:     to,
		Signer: signer,
	}
}

// Route should return the cmname of the module
func (msg MsgSetUnStake) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetUnStake) Type() string { return "set_unstake" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetUnStake) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if msg.To.Empty() {
		return sdk.ErrInvalidAddress(msg.To.String())
	}
	if len(msg.Coins) == 0 {
		return sdk.ErrUnknownRequest("Cannot have no coins")
	}
	for _, coin := range msg.Coins {
		if !coin.IsValid() {
			return sdk.ErrUnknownRequest("Cannot have no invalid coins")
		}
		if !coin.IsZero() {
			return sdk.ErrUnknownRequest("Cannot have a coin with zero value")
		}
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetUnStake) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetUnStake) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
