package observer

import (
	"strings"

	"gitlab.com/thorchain/bepswap/thornode/x/statechain/types"
)

type MockPoolAddressValidator struct {
	poolAddresses types.PoolAddresses
}

func NewMockPoolAddressValidator() *MockPoolAddressValidator {
	return &MockPoolAddressValidator{poolAddresses: types.PoolAddresses{
		Previous: "tbnb1hzwfk6t3sqjfuzlr0ur9lj920gs37gg92gtay9",
		Current:  "tbnb1yycn4mh6ffwpjf584t8lpp7c27ghu03gpvqkfj",
		Next:     "tbnb1hzwfk6t3sqjfuzlr0ur9lj920gs37gg92gtay9",
		RotateAt: 100,
	}}
}
func (mpa *MockPoolAddressValidator) IsValidPoolAddress(addr string) bool {
	if strings.EqualFold(mpa.poolAddresses.Previous.String(), addr) ||
		strings.EqualFold(mpa.poolAddresses.Current.String(), addr) ||
		strings.EqualFold(mpa.poolAddresses.Next.String(), addr) {
		return true
	}
	return false
}
