package types

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

const (
	DefaultCodespace    sdk.CodespaceType = ModuleName
	SuperMajorityFactor uint64            = 3
	MajorityFactor      uint64            = 2
)

// HasSuperMajority return true when it has 2/3 majority
func HasSuperMajority(signers, total int) bool {
	if signers > total {
		return false // will not have majority if THORNode have more signers than node accounts. This shouldn't be possible
	}
	if signers <= 0 {
		return false // edge case
	}
	mU := sdk.NewUint(SuperMajorityFactor)

	// 10*4 / (6.67*2) <= 3
	// 4*4 / (3*2) <= 3
	// 3*4 / (2*2) <= 3
	// Is able to determine "majority" without needing floats or DECs
	tU := sdk.NewUint(uint64(total))
	sU := sdk.NewUint(uint64(signers))
	factor := tU.MulUint64(2).Quo(sU)
	return mU.GTE(factor)
}

// HasMajority return true when it has more than 1/2
func HasSimpleMajority(signers, total int) bool {
	if signers > total {
		return false // will not have majority if THORNode have more signers than node accounts. This shouldn't be possible
	}
	if signers <= 0 {
		return false // edge case
	}
	mU := sdk.NewUint(MajorityFactor)

	// Is able to determine "majority" without needing floats or DECs
	tU := sdk.NewUint(uint64(total))
	sU := sdk.NewUint(uint64(signers))
	factor := tU.Quo(sU)
	return mU.GTE(factor)
}

// GetThreshold calculate threshold
func GetThreshold(value int) (int, error) {
	if value < 0 {
		return 0, errors.New("negative input")
	}
	threshold := int(math.Ceil(float64(value) * 2.0 / 3.0))
	return threshold, nil
}

// ChooseSignerParty use pseodurandom number generate to choose 2/3 majority signer to form a key sign party
func ChooseSignerParty(pubKeys common.PubKeys, seed int64, total int) (common.PubKeys, error) {
	totalCandidates := len(pubKeys)
	signers := common.PubKeys{}
	sort.SliceStable(pubKeys, func(i, j int) bool {
		return pubKeys[i].String() < pubKeys[j].String()
	})

	threshold, err := GetThreshold(total)
	if err != nil {
		return common.PubKeys{}, fmt.Errorf("fail to get threshold: %w", err)
	}
	if totalCandidates < threshold {
		return common.PubKeys{}, fmt.Errorf("total(%d) is less than threshold(%d)", totalCandidates, threshold)
	}
	source := rand.NewSource(seed)
	rnd := rand.New(source)
	for {
		// keep choosing until it get threshold number of signers
		idx := rnd.Intn(totalCandidates)
		k := pubKeys[idx]
		if !signers.Contains(k) {
			signers = append(signers, k)
			if len(signers) == threshold {
				break
			}
		}
	}
	return signers, nil
}
