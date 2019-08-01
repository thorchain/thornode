package swapservice

import (
	"github.com/pkg/errors"
	. "gopkg.in/check.v1"
)

type StakeSuite struct{}

var _ = Suite(&StakeSuite{})

func (s StakeSuite) TestCalculatePoolUnits(c *C) {
	inputs := []struct {
		name         string
		oldPoolUnits float64
		poolRune     float64
		poolToken    float64
		stakeRune    float64
		stakeToken   float64
		poolUnits    float64
		stakerUnits  float64
		expectedErr  error
	}{
		{
			name:         "negative-poolrune",
			oldPoolUnits: 0,
			poolRune:     -100.0,
			expectedErr:  errors.New("negative RUNE in the pool,likely it is corrupted"),
		},
		{
			name:         "negative-pooltoken",
			oldPoolUnits: 0,
			poolRune:     100.0,
			poolToken:    -100.0,
			expectedErr:  errors.New("negative token in the pool,likely it is corrupted"),
		},
		{
			name:         "negative-stakerune",
			oldPoolUnits: 0,
			poolRune:     100.0,
			poolToken:    100.0,
			stakeRune:    -100,
			expectedErr:  errors.New("you can't stake negative rune"),
		},
		{
			name:         "negative-staketoken",
			oldPoolUnits: 0,
			poolRune:     100.0,
			poolToken:    100.0,
			stakeRune:    100,
			stakeToken:   -100,
			expectedErr:  errors.New("you can't stake negative token"),
		},
		{
			name:         "first-stake",
			oldPoolUnits: 0,
			poolRune:     0.0,
			poolToken:    0.0,
			stakeRune:    100,
			stakeToken:   100,
			poolUnits:    100,
			stakerUnits:  100,
			expectedErr:  nil,
		},
		{
			name:         "second-stake",
			oldPoolUnits: 500.0,
			poolRune:     500.0,
			poolToken:    500.0,
			stakeRune:    345,
			stakeToken:   234,
			poolUnits:    787.0168486,
			stakerUnits:  287.016849,
			expectedErr:  nil,
		},
	}

	for _, item := range inputs {
		poolUnits, stakerUnits, err := calculatePoolUnits(item.oldPoolUnits, item.poolRune, item.poolToken, item.stakeRune, item.stakeToken)
		if item.expectedErr == nil {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err.Error(), Equals, item.expectedErr.Error())
		}
		c.Check(round(item.poolUnits), Equals, round(poolUnits))
		c.Check(round(item.stakerUnits), Equals, round(stakerUnits))
	}
}
