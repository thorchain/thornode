package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

type TssKeysignFailVoter struct {
	ID      string           `json:"id"` // checksum of sorted input pubkeys
	Height  int64            `json:"height"`
	Blame   common.Blame     `json:"blame"`
	Signers []sdk.AccAddress `json:"signers"`
}

func NewTssKeysignFailVoter(id string, height int64) TssKeysignFailVoter {
	return TssKeysignFailVoter{
		ID:     id,
		Height: height,
	}
}

// HasSigned - check if given address has signed
func (tss TssKeysignFailVoter) HasSigned(signer sdk.AccAddress) bool {
	for _, sign := range tss.Signers {
		if sign.Equals(signer) {
			return true
		}
	}
	return false
}

// Sign this voter with given signer address
func (tss *TssKeysignFailVoter) Sign(signer sdk.AccAddress) bool {
	if tss.HasSigned(signer) {
		return false
	}
	tss.Signers = append(tss.Signers, signer)
	return true
}

// Determine if this tss pool has enough signers
func (tss *TssKeysignFailVoter) HasConsensus(nas NodeAccounts) bool {
	var count int
	for _, signer := range tss.Signers {
		if nas.IsNodeKeys(signer) {
			count += 1
		}
	}
	if HasSimpleMajority(count, len(nas)) {
		return true
	}

	return false
}

// Empty to check whether this Voter is empty or not
func (tss *TssKeysignFailVoter) Empty() bool {
	return len(tss.ID) == 0 || tss.Height == 0
}

func (tss *TssKeysignFailVoter) String() string {
	return tss.ID
}
