package pubkeymanager

import (
	"errors"
	"fmt"
	"strings"

	"gitlab.com/thorchain/thornode/common"
)

var (
	previous = "tbnb1hzwfk6t3sqjfuzlr0ur9lj920gs37gg92gtay9"
	current  = "tbnb1yycn4mh6ffwpjf584t8lpp7c27ghu03gpvqkfj"
	next     = "tbnb1hzwfk6t3sqjfuzlr0ur9lj920gs37gg92gtay9"
	top      = "tbnb186nvjtqk4kkea3f8a30xh4vqtkrlu2rm9xgly3"
	validpb  = "thorpub1addwnpepqfgfxharps79pqv8fv9ndqh90smw8c3slrtrssn58ryc5g3p9sx856x07yn"
)

type MockPoolAddressValidator struct{}

func NewMockPoolAddressValidator() *MockPoolAddressValidator {
	return &MockPoolAddressValidator{}
}

func matchTestAddress(addr, testAddr string, chain common.Chain) (bool, common.ChainPoolInfo) {
	if strings.EqualFold(testAddr, addr) {
		pubKey, _ := common.NewPubKey(validpb)
		cpi, err := common.NewChainPoolInfo(chain, pubKey)
		cpi.PoolAddress = common.Address(testAddr)
		fmt.Println(err)
		return true, cpi
	}
	return false, common.EmptyChainPoolInfo
}

func (mpa *MockPoolAddressValidator) FetchPubKeys()              {}
func (mpa *MockPoolAddressValidator) GetPubKeys() common.PubKeys { return nil }
func (mpa *MockPoolAddressValidator) GetSignPubKeys() common.PubKeys {
	pubKey, _ := common.NewPubKey(validpb)
	return common.PubKeys{pubKey}
}
func (mpa *MockPoolAddressValidator) GetNodePubKey() common.PubKey       { return common.EmptyPubKey }
func (mpa *MockPoolAddressValidator) HasPubKey(pk common.PubKey) bool    { return false }
func (mpa *MockPoolAddressValidator) AddPubKey(pk common.PubKey, _ bool) {}
func (mpa *MockPoolAddressValidator) AddNodePubKey(pk common.PubKey)     {}
func (mpa *MockPoolAddressValidator) RemovePubKey(pk common.PubKey)      {}
func (mpa *MockPoolAddressValidator) Start() error                       { return errors.New("Kaboom!") }
func (mpa *MockPoolAddressValidator) Stop() error                        { return errors.New("Kaboom!") }

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
	matchTop, cpi := matchTestAddress(addr, top, chain)
	if matchTop {
		return matchTop, cpi
	}
	return false, common.EmptyChainPoolInfo
}
