package swapservice

import (
	"github.com/pkg/errors"
	. "gopkg.in/check.v1"
)

type UnstakeSuite struct{}

var _ = Suite(&UnstakeSuite{})

func (s UnstakeSuite) TestCalculateUnsake(c *C) {
	inputs := []struct {
		name                  string
		poolUnit              float64
		poolRune              float64
		poolToken             float64
		stakerUnit            float64
		percentage            float64
		expectedWithdrawRune  float64
		expectedWithdrawToken float64
		expectedUnitLeft      float64
		expectedErr           error
	}{
		{
			name:        "zero-poolunit",
			poolUnit:    0,
			expectedErr: errors.New("poolUnits can't be zero or negative"),
		},
		{
			name:        "negative-poolunit",
			poolUnit:    -100,
			expectedErr: errors.New("poolUnits can't be zero or negative"),
		},
		{
			name:        "zero-poolrune",
			poolUnit:    500,
			expectedErr: errors.New("pool rune balance can't be zero or negative"),
		},
		{
			name:        "negative-poolrune",
			poolUnit:    500,
			poolRune:    -100,
			expectedErr: errors.New("pool rune balance can't be zero or negative"),
		},
		{
			name:        "zero-pooltoken",
			poolUnit:    500,
			poolRune:    500,
			poolToken:   0,
			expectedErr: errors.New("pool token balance can't be zero or negative"),
		},
		{
			name:        "negative-poolrune",
			poolUnit:    500,
			poolRune:    500,
			poolToken:   -100,
			expectedErr: errors.New("pool token balance can't be zero or negative"),
		},
		{
			name:        "negative-stakerUnit",
			poolUnit:    500,
			poolRune:    500,
			poolToken:   5100,
			stakerUnit:  -100,
			expectedErr: errors.New("staker unit can't be negative"),
		},
		{
			name:        "negative-percentage",
			poolUnit:    500,
			poolRune:    500,
			poolToken:   500,
			stakerUnit:  100,
			percentage:  -20,
			expectedErr: errors.Errorf("percentage %f is not valid", -20.0),
		},
		{
			name:        "percentage-larger-than-100",
			poolUnit:    500,
			poolRune:    500,
			poolToken:   500,
			stakerUnit:  100,
			percentage:  120,
			expectedErr: errors.Errorf("percentage %f is not valid", 120.0),
		},
		{
			name:                  "unstake-1",
			poolUnit:              700,
			poolRune:              700,
			poolToken:             700,
			stakerUnit:            200,
			percentage:            100,
			expectedUnitLeft:      0,
			expectedWithdrawToken: 200,
			expectedWithdrawRune:  200,
			expectedErr:           nil,
		},
		// TOOD add more cases in
	}

	for _, item := range inputs {
		withDrawRune, withDrawToken, unitAfter, err := calculateUnstake(item.poolUnit, item.poolRune, item.poolToken, item.stakerUnit, item.percentage)
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}
		c.Check(round(item.expectedWithdrawRune), Equals, withDrawRune)
		c.Check(round(item.expectedWithdrawToken), Equals, withDrawToken)
		c.Check(round(item.expectedUnitLeft), Equals, unitAfter)
	}
}
