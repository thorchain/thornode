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
	Chains      common.Chains    `json:"chains"`
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
func (tss *TssVoter) Sign(signer sdk.AccAddress, chains common.Chains) bool {
	if tss.HasSigned(signer) {
		return false
	}
	for _, pk := range tss.PubKeys {
		addr, err := pk.GetThorAddress()
		if addr.Equals(signer) && err == nil {
			tss.Signers = append(tss.Signers, signer)
			tss.Chains = append(tss.Chains, chains...)
			return true
		}
	}
	return false
}

// ConsensusChains - get a list o chains that have 2/3rds majority
func (tss *TssVoter) ConsensusChains() common.Chains {
	chainCount := make(map[common.Chain]int, 0)
	for _, chain := range tss.Chains {
		if _, ok := chainCount[chain]; !ok {
			chainCount[chain] = 0
		}
		chainCount[chain]++
	}

	chains := make(common.Chains, 0)
	for chain, count := range chainCount {
		if HasSuperMajority(count, len(tss.PubKeys)) {
			chains = append(chains, chain)
		}
	}

	return chains
}

// Determine if this tss pool has enough signers
func (tss *TssVoter) HasConsensus() bool {
	if HasSuperMajority(len(tss.Signers), len(tss.PubKeys)) {
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
