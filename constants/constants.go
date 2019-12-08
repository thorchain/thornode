package constants

import (
	"log"

	"github.com/blang/semver"
)

// The version of this software
var SWVersion = semver.MustParse("0.1.0")

type Constants struct {
	// The emission curve targets a ~2% emission after 10 years (similar to Bitcoin).
	// The BlocksPerYear directly affects emission rate, and may be updated if markedly different in production
	// Day 0 Emission is ~25%, Year 1 Emission is ~20%, Year 10 Emission is ~2%
	EmissionCurve  int // An arbitrary factor to target desired curve
	BlocksPerYear  int // Number of blocks per year
	TransactionFee int // A Rune fee on all swaps and withdrawals

	// A new pool is enabled on a cycle
	NewPoolCycle            int64 // Enable a pool every X blocks
	MinmumNodesForYggdrasil int   // No yggdrasil pools if THORNode have less than X active nodes
	MinmumNodesForBFT       int   // Minimum node count to keep network running. Below this, Ragnarök is performed.

	// validator rotation
	ValidatorRotateInNumBeforeFull  int // How many validators should THORNode nominate before THORNode reach the desire validator set
	ValidatorRotateOutNumBeforeFull int // How many validators should THORNode queued to be rotate out before THORNode reach the desire validator set)
	ValidatorRotateNumAfterFull     int // How many validators should THORNode nominate after THORNode reach the desire validator set
	DesireValidatorSet              int // desire validator set
	RotatePerBlockHeight            int // How many blocks THORNode try to rotate validators
	ValidatorsChangeWindow          int // When should THORNode open the rotate window, nominate validators, and identify who should be out
	LeaveProcessPerBlockHeight      int // after how many blocks THORNode will process leave queue

	// Slashing
	LackOfObservationPenalty int64  // add two slash point for each block where a node does not observe
	SigningTransactionPeriod int64  // how many blocks before a request to sign a tx by yggdrasil pool, is counted as delinquent.
	MinimumBondInRune        uint64 // minimum bond amount to become a validator
}

func GetConstants(ver semver.Version) Constants {
	if ver.GTE(semver.MustParse("0.1.0")) {
		return Constants{
			EmissionCurve:                   6,               // An arbitrary factor to target desired curve
			BlocksPerYear:                   6311390,         // (365.2425 * 86400) / (Seconds per THORChain block) -> 31556952 / 5 -> 6311390
			TransactionFee:                  100000000,       // 1.0 Rune
			NewPoolCycle:                    50000,           // Every 50,00 blok (~3 days)
			MinmumNodesForYggdrasil:         6,               // No yggdrasil pools if THORNode have less than 6 active nodes
			MinmumNodesForBFT:               4,               // Minimum node count to keep network running. Below this, Ragnarök is performed.
			ValidatorRotateInNumBeforeFull:  2,               // How many validators should THORNode nominate before THORNode reach the desire validator set
			ValidatorRotateOutNumBeforeFull: 1,               // How many validators should THORNode queued to be rotate out before THORNode reach the desire validator set)
			ValidatorRotateNumAfterFull:     1,               // How many validators should THORNode nominate after THORNode reach the desire validator set
			DesireValidatorSet:              33,              // desire validator set
			RotatePerBlockHeight:            17280,           // How many blocks THORNode try to rotate validators
			ValidatorsChangeWindow:          1200,            // When should THORNode open the rotate window, nominate validators, and identify who should be out
			LeaveProcessPerBlockHeight:      4320,            // after how many blocks THORNode will process leave queue
			LackOfObservationPenalty:        2,               // add two slash point for each block where a node does not observe
			SigningTransactionPeriod:        100,             // how many blocks before a request to sign a tx by yggdrasil pool, is counted as delinquent.
			MinimumBondInRune:               100000000000000, // 1 million rune
		}
	}
	log.Fatal("Unable to determine contstants")
	return Constants{}
}
