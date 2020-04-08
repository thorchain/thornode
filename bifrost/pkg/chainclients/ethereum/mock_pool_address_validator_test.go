package ethereum

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/common"
)

var (
	previous = "0x8bc3da587def887b5c822105729ee1d6af05a5ca"
	current  = "0x798abda6cc246d0edba912092a2a3dbd3d11191b"
	next     = "0x9750ed9f7a71df35d337c0667e53c7f005e2c13c"
	top      = "0xc989de8af3f45241ae2969a191770ae367200977"
	validpb  = "thorpub1addwnpepqfgfxharps79pqv8fv9ndqh90smw8c3slrtrssn58ryc5g3p9sx856x07yn"
)

type MockPoolAddressValidator struct {
}

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
