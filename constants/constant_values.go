package constants

import (
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
	ValidatorsChangeWindow
	LeaveProcessPerBlockHeight
	BadValidatorRate
	OldValidatorRate
	LackOfObservationPenalty
	SigningTransactionPeriod
	MinimumBondInRune
	FundMigrationInterval
	WhiteListGasAsset
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
	ValidatorsChangeWindow:          "ValidatorsChangeWindow",
	LeaveProcessPerBlockHeight:      "LeaveProcessPerBlockHeight",
	BadValidatorRate:                "BadValidatorRate",
	OldValidatorRate:                "OldValidatorRate",
	LackOfObservationPenalty:        "LackOfObservationPenalty",
	SigningTransactionPeriod:        "SigningTransactionPeriod",
	MinimumBondInRune:               "MinimumBondInRune",
	FundMigrationInterval:           "FundMigrationInterval",
	WhiteListGasAsset:               "WhiteListGasAsset",
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
	GetInt64Value(name ConstantName) int64
}

// GetConstantValues will return an  implementation of ConstantValues which provide ways to get constant values
func GetConstantValues(ver semver.Version) ConstantValues {
	if ver.GTE(SWVersion) {
		return NewConstantValue010()
	}
	return nil
}
