package constants

import (
	"github.com/blang/semver"
	. "gopkg.in/check.v1"
)

type ConstantsTestSuite struct{}

var _ = Suite(&ConstantsTestSuite{})

func (ConstantsTestSuite) TestConstantName_String(c *C) {
	constantNames := []ConstantName{
		EmissionCurve,
		BlocksPerYear,
		TransactionFee,
		NewPoolCycle,
		MinimumNodesForYggdrasil,
		MinimumNodesForBFT,
		ValidatorRotateInNumBeforeFull,
		ValidatorRotateOutNumBeforeFull,
		ValidatorRotateNumAfterFull,
		DesireValidatorSet,
		RotatePerBlockHeight,
		ValidatorsChangeWindow,
		LeaveProcessPerBlockHeight,
		BadValidatorRate,
		OldValidatorRate,
		LackOfObservationPenalty,
		SigningTransactionPeriod,
		DoubleSignMaxAge,
		MinimumBondInRune,
	}
	for _, item := range constantNames {
		c.Assert(item.String(), Not(Equals), "NA")
	}
}

func (ConstantsTestSuite) TestGetConstantValues(c *C) {
	ver := semver.MustParse("0.0.9")
	c.Assert(GetConstantValues(ver), IsNil)
	c.Assert(GetConstantValues(SWVersion), NotNil)
}
