package types

import (
	"crypto/sha256"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

// MsgTssPool defines a MsgTssPool message
type MsgTssPool struct {
	ID         string          `json:"id"`
	PoolPubKey common.PubKey   `json:"pool_pub_key"`
	PubKeys    []common.PubKey `json:"pubkeys"`
	Signer     sdk.AccAddress  `json:"signer"`
}

// NewMsgTssPool is a constructor function for MsgTssPool
func NewMsgTssPool(pks []common.PubKey, poolpk common.PubKey, signer sdk.AccAddress) MsgTssPool {

	// ensure input pubkeys list is deterministically sorted
	sort.Slice(pks, func(i, j int) bool {
		return pks[i].String() < pks[j].String()
	})

	// get the checksum of our input pubkeys, as our identifier
	var buf []byte
	for _, pk := range pks {
		buf = append(buf, []byte(pk)...)
	}
	id := fmt.Sprintf("%X", sha256.Sum256(buf))

	return MsgTssPool{
		ID:         id,
		PubKeys:    pks,
		PoolPubKey: poolpk,
		Signer:     signer,
	}
}

// Route should return the cmname of the module
func (msg MsgTssPool) Route() string { return RouterKey }

// Type should return the action
func (msg MsgTssPool) Type() string { return "set_tss_pool" }

// ValidateBasic runs stateless checks on the message
func (msg MsgTssPool) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress(msg.Signer.String())
	}
	if len(msg.ID) == 0 {
		return sdk.ErrUnknownRequest("ID cannot be blank")
	}
	if len(msg.PubKeys) < 2 {
		return sdk.ErrUnknownRequest("Must have at least 2 pub keys")
	}
	for _, pk := range msg.PubKeys {
		if pk.IsEmpty() {
			return sdk.ErrUnknownRequest("Pubkey cannot be empty")
		}
	}
	if msg.PoolPubKey.IsEmpty() {
		return sdk.ErrUnknownRequest("Pool pubkey cannot be empty")
	}
	// ensure pool pubkey is a valid bech32 pubkey
	if _, err := common.NewPubKey(msg.PoolPubKey.String()); err != nil {
		return sdk.ErrUnknownRequest(err.Error())
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgTssPool) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgTssPool) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
