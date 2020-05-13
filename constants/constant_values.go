package constants

import (
	"fmt"

	"github.com/blang/semver"
)

// ConstantName the name we used to get constant values
type ConstantName int

const (
	EmissionCurve ConstantName = iota
	BlocksPerYear
	TransactionFee
	NewPoolCycle
	MinimumNodesForYggdrasil
	MinimumNodesForBFT
	ValidatorRotateInNumBeforeFull
	ValidatorRotateOutNumBeforeFull
	ValidatorRotateNumAfterFull
	DesireValidatorSet
	RotatePerBlockHeight
	RotateRetryBlocks
	ValidatorsChangeWindow
	LeaveProcessPerBlockHeight
	BadValidatorRate
	OldValidatorRate
	LackOfObservationPenalty
	SigningTransactionPeriod
	DoubleSignMaxAge
	MinimumBondInRune
	FundMigrationInterval
	WhiteListGasAsset
	ArtificialRagnarokBlockHeight
	MaximumStakeRune
	StrictBondStakeRatio
	DefaultPoolStatus
	FailKeygenSlashPoints
	FailKeySignSlashPoints
	StakeLockUpBlocks
	ObserveSlashPoints
)

var nameToString = map[ConstantName]string{
	EmissionCurve:                   "EmissionCurve",
	BlocksPerYear:                   "BlockPerYear",
	TransactionFee:                  "TransactionFee",
	NewPoolCycle:                    "NewPoolCycle",
	MinimumNodesForYggdrasil:        "MinimumNodesForYggdrasil",
	MinimumNodesForBFT:              "MinimumNodesForBFT",
	ValidatorRotateInNumBeforeFull:  "ValidatorRotateInNumBeforeFull",
	ValidatorRotateOutNumBeforeFull: "ValidatorRotateOutNumBeforeFull",
	ValidatorRotateNumAfterFull:     "ValidatorRotateNumAfterFull",
	DesireValidatorSet:              "DesireValidatorSet",
	RotatePerBlockHeight:            "RotatePerBlockHeight",
	RotateRetryBlocks:               "RotateRetryBlocks",
	ValidatorsChangeWindow:          "ValidatorsChangeWindow",
	LeaveProcessPerBlockHeight:      "LeaveProcessPerBlockHeight",
	BadValidatorRate:                "BadValidatorRate",
	OldValidatorRate:                "OldValidatorRate",
	LackOfObservationPenalty:        "LackOfObservationPenalty",
	SigningTransactionPeriod:        "SigningTransactionPeriod",
	DoubleSignMaxAge:                "DoubleSignMaxAge",
	MinimumBondInRune:               "MinimumBondInRune",
	FundMigrationInterval:           "FundMigrationInterval",
	WhiteListGasAsset:               "WhiteListGasAsset",
	ArtificialRagnarokBlockHeight:   "ArtificialRagnarokBlockHeight",
	MaximumStakeRune:                "MaximumStakeRune",
	StrictBondStakeRatio:            "StrictBondStakeRatio",
	DefaultPoolStatus:               "DefaultPoolStatus",
	FailKeygenSlashPoints:           "FailKeygenSlashPoints",
	FailKeySignSlashPoints:          "FailKeySignSlashPoints",
	StakeLockUpBlocks:               "StakeLockUpBlocks",
	ObserveSlashPoints:              "ObserveSlashPoints",
}

// String implement fmt.stringer
func (cn ConstantName) String() string {
	val, ok := nameToString[cn]
	if !ok {
		return "NA"
	}
	return val
}

// ConstantValues define methods used to get constant values
type ConstantValues interface {
	fmt.Stringer
	GetInt64Value(name ConstantName) int64
	GetBoolValue(name ConstantName) bool
	GetStringValue(name ConstantName) string
}

// GetConstantValues will return an  implementation of ConstantValues which provide ways to get constant values
func GetConstantValues(ver semver.Version) ConstantValues {
	if ver.GTE(semver.MustParse("0.1.0")) {
		return NewConstantValue010()
	}
	return nil
}
