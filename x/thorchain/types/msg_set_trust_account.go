package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

// MsgSetTrustAccount defines a MsgSetTrustAccount message
type MsgSetTrustAccount struct {
	NodePubKeys         common.PubKeys `json:"node_pub_keys"`
	ValidatorConsPubKey string         `json:"validator_cons_pub_key"`
	Signer              sdk.AccAddress `json:"signer"`
}

// NewMsgSetTrustAccount is a constructor function for NewMsgAddTrustAccount
func NewMsgSetTrustAccount(nodePubKeys common.PubKeys, validatorConsPubKey string, signer sdk.AccAddress) MsgSetTrustAccount {
	return MsgSetTrustAccount{
		NodePubKeys:         nodePubKeys,
		ValidatorConsPubKey: validatorConsPubKey,
		Signer:              signer,
	}
}

// Route should return the cmname of the module
func (msg MsgSetTrustAccount) Route() string { return RouterKey }

// Type should return the action
func (msg MsgSetTrustAccount) Type() string { return "set_trust_account" }

// ValidateBasic runs stateless checks on the message
func (msg MsgSetTrustAccount) ValidateBasic() sdk.Error {
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
func (msg MsgSetTrustAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgSetTrustAccount) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
