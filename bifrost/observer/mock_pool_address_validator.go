package observer

import (
	"fmt"
	"strings"

	"gitlab.com/thorchain/bepswap/thornode/common"
	"gitlab.com/thorchain/bepswap/thornode/x/thorchain/types"
)

var (
	previous = "tbnb1hzwfk6t3sqjfuzlr0ur9lj920gs37gg92gtay9"
	current  = "tbnb1yycn4mh6ffwpjf584t8lpp7c27ghu03gpvqkfj"
	next     = "tbnb1hzwfk6t3sqjfuzlr0ur9lj920gs37gg92gtay9"
)

type MockPoolAddressValidator struct {
	poolAddresses types.PoolAddresses
}

func NewMockPoolAddressValidator() *MockPoolAddressValidator {
	return &MockPoolAddressValidator{}
}

func matchTestAddress(addr, testAddr string, chain common.Chain) (bool, common.ChainPoolInfo) {
	if strings.EqualFold(testAddr, addr) {
		cpi, err := common.NewChainPoolInfo(chain, types.GetRandomPubKey())
		cpi.PoolAddress = common.Address(testAddr)
		fmt.Println(err)
		return true, cpi
	}
	return false, common.EmptyChainPoolInfo
}

func (mpa *MockPoolAddressValidator) AddPubKey(pk common.PubKeys)    {}
func (mpa *MockPoolAddressValidator) RemovePubKey(pk common.PubKeys) {}

func (mpa *MockPoolAddressValidator) IsValidAddress(addr string, chain common.Chain) bool {
	ok, _ := mpa.IsValidPoolAddress(addr, chain)
	return ok
}

func (mpa *MockPoolAddressValidator) IsValidPoolAddress(addr string, chain common.Chain) (bool, common.ChainPoolInfo) {
	matchCurrent, cpi := matchTestAddress(addr, current, chain)
	if matchCurrent {
		return matchCurrent, cpi
	}
	matchPrevious, cpi := matchTestAddress(addr, previous, chain)
	if matchPrevious {
		return matchPrevious, cpi
	}
	matchNext, cpi := matchTestAddress(addr, next, chain)
	if matchNext {
		return matchNext, cpi
	}
	return false, common.EmptyChainPoolInfo

}
