package swapservice

import (
	"testing"

	"github.com/pkg/errors"
)

func TestCalculatePoolUnits(t *testing.T) {
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
		t.Run(item.name, func(st *testing.T) {
			poolUnits, stakerUnits, err := calculatePoolUnits(item.oldPoolUnits, item.poolRune, item.poolToken, item.stakeRune, item.stakeToken)
			if nil != item.expectedErr {
				if nil == err {
					t.Errorf("we are expecting %s, however we got nil", item.expectedErr)
					return
				}
				if item.expectedErr.Error() != err.Error() {
					t.Errorf("we are expecting %s, however we got %s", item.expectedErr, err)
					return
				}
				return
			}
			if err != nil {
				t.Errorf("we are not expecting err, however we got %s", err)
				return
			}
			t.Log("poolunits", poolUnits)
			if round(item.poolUnits) != round(poolUnits) {
				t.Errorf("we are expecting poolUnits to be %f however we got %f ", item.poolUnits, poolUnits)
				return
			}
			if round(item.stakerUnits) != round(stakerUnits) {
				t.Errorf("we are expecting staker units to be %f however we got %f ", item.stakerUnits, stakerUnits)
				return
			}
		})
	}
}
