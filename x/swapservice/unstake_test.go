package swapservice

import (
	"testing"

	"github.com/pkg/errors"
)

func TestCalculateUnsake(t *testing.T) {
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
		t.Run(item.name, func(st *testing.T) {
			withDrawRune, withDrawToken, unitAfter, err := calculateUnsake(item.poolUnit, item.poolRune, item.poolToken, item.stakerUnit, item.percentage)
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
			if round(item.expectedWithdrawRune) != withDrawRune {
				t.Errorf("expected withdraw rune is : %f, however we got : %f", item.expectedWithdrawRune, withDrawRune)
				return
			}
			if round(item.expectedWithdrawToken) != withDrawToken {
				t.Errorf("expected withdraw token is : %f, however we got : %f", item.expectedWithdrawToken, withDrawToken)
				return
			}
			if round(item.expectedUnitLeft) != unitAfter {
				t.Errorf("expected poolunit after withdraw is : %f, however we got : %f", item.expectedUnitLeft, unitAfter)
				return
			}
		})
	}
}
