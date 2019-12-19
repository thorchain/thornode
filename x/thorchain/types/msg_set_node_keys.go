package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// MsgSetNodeKeys defines a MsgSetNodeKeys message
type MsgSetNodeKeys struct {
	NodePubKeys         common.PubKeys `json:"node_pub_keys"`
	ValidatorConsPubKey string         `json:"validator_cons_pub_key"`
	Signer              sdk.AccAddress `json:"signer"`
}

// NewMsgSetNodeKeys is a constructor function for NewMsgAddNodeKeys
func NewMsgSetNodeKeys(nodePubKeys common.PubKeys, validatorConsPubKey string, signer sdk.AccAddress) MsgSetNodeKeys {
	return MsgSetNodeKeys{
		NodePubKeys:         nodePubKeys,
		ValidatorConsPubKey: validatorConsPubKey,
		Signer:              signer,
	}
}

// Route should return the cmname of the module
func (msg MsgSetNodeKeys) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetNodeKeys) Type() string { return "set_node_keys" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetNodeKeys) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if len(msg.ValidatorConsPubKey) == 0 {
		return sdk.ErrUnknownRequest("validator consensus pubkey cannot be empty")
	}
	if msg.NodePubKeys.IsEmpty() {
		return sdk.ErrUnknownRequest("node pub keys cannot be empty")
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgSetNodeKeys) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetNodeKeys) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
