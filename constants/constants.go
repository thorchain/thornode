package constants

import (
	"github.com/blang/semver"
)

// The version of this software
var SWVersion, _ = semver.Make("0.1.0")

// The emission curve targets a ~2% emission after 10 years (similar to Bitcoin).
// The BlocksPerYear directly affects emission rate, and may be updated if markedly different in production
// Day 0 Emission is ~25%, Year 1 Emission is ~20%, Year 10 Emission is ~2%
const EmissionCurve = 6          // An arbitrary factor to target desired curve
const BlocksPerYear = 6311390    // (365.2425 * 86400) / (Seconds per THORChain block) -> 31556952 / 5 -> 6311390
const TransactionFee = 100000000 // A 1.0 Rune fee on all swaps and withdrawals

// A new pool is enabled on a cycle
const NewPoolCycle = 50000        // Enable a pool every 50,000 blocks (~3 days)
const MinmumNodesForYggdrasil = 6 // No yggdrasil pools if THORNode have less than 6 active nodes
const MinmumNodesForBFT = 4       // Minimum node count to keep network running. Below this, Ragnarök is performed.

// validator rotation
const (
	ValidatorRotateInNumBeforeFull  = 2     // How many validators should THORNode nominate before THORNode reach the desire validator set
	ValidatorRotateOutNumBeforeFull = 1     // How many validators should THORNode queued to be rotate out before THORNode reach the desire validator set)
	ValidatorRotateNumAfterFull     = 1     // How many validators should THORNode nominate after THORNode reach the desire validator set
	DesireValidatorSet              = 33    // desire validator set
	RotatePerBlockHeight            = 17280 // How many blocks THORNode try to rotate validators
	ValidatorsChangeWindow          = 1200  // When should THORNode open the rotate window, nominate validators, and identify who should be out
	LeaveProcessPerBlockHeight      = 4320  // after how many blocks THORNode will process leave queue
)

// Slashing
const LackOfObservationPenalty int64 = 2   // add two slash point for each block where a node does not observe
const SigningTransactionPeriod int64 = 100 // how many blocks before a request to sign a tx by yggdrasil pool, is counted as delinquent.
