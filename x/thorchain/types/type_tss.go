package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
)

type TssVoter struct {
	ID          string           `json:"id"` // checksum of sorted input pubkeys
	PoolPubKey  common.PubKey    `json:"pool_pub_key"`
	PubKeys     common.PubKeys   `json:"pubkeys"`
	BlockHeight int64            `json:"block_height"`
	Signers     []sdk.AccAddress `json:"signers"`
}

func NewTssVoter(id string, pks common.PubKeys, pool common.PubKey) TssVoter {
	return TssVoter{
		ID:         id,
		PubKeys:    pks,
		PoolPubKey: pool,
	}
}

// HasSigned - check if given address has signed
func (tss TssVoter) HasSigned(signer sdk.AccAddress) bool {
	for _, sign := range tss.Signers {
		if sign.Equals(signer) {
			return true
		}
	}
	return false
}

// Sign this voter with given signer address
func (tss *TssVoter) Sign(signer sdk.AccAddress) {
	if !tss.HasSigned(signer) {
		tss.Signers = append(tss.Signers, signer)
	}
}

// Determine if this tss pool has enough signers
func (tss *TssVoter) HasConensus(nas NodeAccounts) bool {
	var count int
	for _, signer := range tss.Signers {
		if nas.IsNodeKeys(signer) {
			count += 1
		}
	}
	if HasMajority(count, len(nas)) {
		return true
	}

	return false
}

func (tss *TssVoter) Empty() bool {
	if len(tss.ID) == 0 || len(tss.PoolPubKey) == 0 || len(tss.PubKeys) == 0 {
		return true
	}
	return false
}

func (tss *TssVoter) String() string {
	return tss.ID
}
