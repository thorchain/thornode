// Package constants  contains all the constants used by thorchain
// by default all the settings in this is for mainnet
package constants

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blang/semver"
)

var (
	GitCommit       string // sha1 revision used to build the program
	BuildTime       string // when the executable was built
	Version         string // software version
	int64Overrides  = map[ConstantName]int64{}
	boolOverrides   = map[ConstantName]bool{}
	stringOverrides = map[ConstantName]string{}
)

// The version of this software
var SWVersion, _ = semver.Make(Version)

// ConstantValue010 implement ConstantValues interface for version 0.1.0
type ConstantValue010 struct {
	int64values  map[ConstantName]int64
	boolValues   map[ConstantName]bool
	stringValues map[ConstantName]string
}

// NewConstantValue010 get new instance of ConstantValue010
func NewConstantValue010() *ConstantValue010 {
	return &ConstantValue010{
		int64values: map[ConstantName]int64{
			EmissionCurve:                   6,
			BlocksPerYear:                   6311390,
			TransactionFee:                  100_000_000,         // A 1.0 Rune fee on all swaps and withdrawals
			NewPoolCycle:                    50000,               // Enable a pool every 50,000 blocks (~3 days)
			MinimumNodesForYggdrasil:        6,                   // No yggdrasil pools if THORNode have less than 6 active nodes
			MinimumNodesForBFT:              4,                   // Minimum node count to keep network running. Below this, Ragnarök is performed.
			ValidatorRotateInNumBeforeFull:  2,                   // How many validators should THORNode nominate before THORNode reach the desire validator set
			ValidatorRotateOutNumBeforeFull: 1,                   // How many validators should THORNode queued to be rotate out before THORNode reach the desire validator set)
			ValidatorRotateNumAfterFull:     1,                   // How many validators should THORNode nominate after THORNode reach the desire validator set
			DesireValidatorSet:              33,                  // desire validator set
			FundMigrationInterval:           360,                 // number of blocks THORNode will attempt to move funds from a retiring vault to an active one
			RotatePerBlockHeight:            51840,               // How many blocks THORNode try to rotate validators
			RotateRetryBlocks:               720,                 // How many blocks until we retry a churn (only if we haven't had a successful churn in RotatePerBlockHeight blocks
			BadValidatorRate:                51840,               // rate to mark a validator to be rotated out for bad behavior
			OldValidatorRate:                51840,               // rate to mark a validator to be rotated out for age
			LackOfObservationPenalty:        2,                   // add two slash point for each block where a node does not observe
			SigningTransactionPeriod:        300,                 // how many blocks before a request to sign a tx by yggdrasil pool, is counted as delinquent.
			DoubleSignMaxAge:                24,                  // number of blocks to limit double signing a block
			MinimumBondInRune:               100_000_000_000_000, // 1 million rune
			WhiteListGasAsset:               1000,                // thor coins we will be given to the validator
			FailKeygenSlashPoints:           720,                 // slash for 720 blocks , which equals 1 hour
			FailKeySignSlashPoints:          2,                   // slash for 2 blocks
			StakeLockUpBlocks:               17280,               // the number of blocks staker can unstake after their stake
		},
		boolValues: map[ConstantName]bool{
			StrictBondStakeRatio: true,
		},
		stringValues: map[ConstantName]string{
			DefaultPoolStatus: "Bootstrap",
		},
	}
}

// GetInt64Value get value in int64 type, if it doesn't exist then it will return the default value of int64, which is 0
func (cv *ConstantValue010) GetInt64Value(name ConstantName) int64 {
	// check overrides first
	v, ok := int64Overrides[name]
	if ok {
		return v
	}

	v, ok = cv.int64values[name]
	if !ok {
		return 0
	}
	return v
}

// GetBoolValue retrieve a bool constant value from the map
func (cv *ConstantValue010) GetBoolValue(name ConstantName) bool {
	v, ok := boolOverrides[name]
	if ok {
		return v
	}
	v, ok = cv.boolValues[name]
	if !ok {
		return false
	}
	return v
}

// GetStringValue retrieve a string const value from the map
func (cv *ConstantValue010) GetStringValue(name ConstantName) string {
	v, ok := stringOverrides[name]
	if ok {
		return v
	}
	v, ok = cv.stringValues[name]
	if ok {
		return v
	}
	return ""
}

func (cv *ConstantValue010) String() string {
	sb := strings.Builder{}
	for k, v := range cv.int64values {
		if overrideValue, ok := int64Overrides[k]; ok {
			sb.WriteString(fmt.Sprintf("%s:%d\n", k, overrideValue))
			continue
		}
		sb.WriteString(fmt.Sprintf("%s:%d\n", k, v))
	}
	for k, v := range cv.boolValues {
		if overrideValue, ok := boolOverrides[k]; ok {
			sb.WriteString(fmt.Sprintf("%s:%v\n", k, overrideValue))
			continue
		}
		sb.WriteString(fmt.Sprintf("%s:%v\n", k, v))
	}
	return sb.String()
}

// MarshalJSON marshal result to json format
func (cv ConstantValue010) MarshalJSON() ([]byte, error) {
	var result struct {
		Int64Values  map[string]int64  `json:"int_64_values"`
		BoolValues   map[string]bool   `json:"bool_values"`
		StringValues map[string]string `json:"string_values"`
	}
	result.Int64Values = make(map[string]int64)
	result.BoolValues = make(map[string]bool)
	result.StringValues = make(map[string]string)
	for k, v := range cv.int64values {
		result.Int64Values[k.String()] = v
	}
	for k, v := range int64Overrides {
		result.Int64Values[k.String()] = v
	}
	for k, v := range cv.boolValues {
		result.BoolValues[k.String()] = v
	}
	for k, v := range boolOverrides {
		result.BoolValues[k.String()] = v
	}
	for k, v := range cv.stringValues {
		result.StringValues[k.String()] = v
	}
	for k, v := range stringOverrides {
		result.StringValues[k.String()] = v
	}

	return json.MarshalIndent(result, "", "	")
}
