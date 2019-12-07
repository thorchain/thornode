package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// PoolAddressDummyMgr is going to manage the pool addresses , rotate etc
type PoolAddressDummyMgr struct {
	currentPoolAddresses       *PoolAddresses
	observedNextPoolAddrPubKey common.PoolPubKeys
	isRotateWindowOpen         bool
}

// NewPoolAddressDummyMgr create a new PoolAddressDummyMgr
func NewPoolAddressDummyMgr() *PoolAddressDummyMgr {
	addrs := NewPoolAddresses(GetRandomPoolPubKeys(), GetRandomPoolPubKeys(), GetRandomPoolPubKeys(), 500, 100)
	return &PoolAddressDummyMgr{
		currentPoolAddresses: addrs,
	}
}

// GetCurrentPoolAddresses return current pool addresses
func (pm *PoolAddressDummyMgr) GetCurrentPoolAddresses() *PoolAddresses {
	return pm.currentPoolAddresses
}

func (pm *PoolAddressDummyMgr) IsRotateWindowOpen() bool {
	return pm.isRotateWindowOpen
}

func (pm *PoolAddressDummyMgr) ObservedNextPoolAddrPubKey() common.PoolPubKeys {
	return pm.observedNextPoolAddrPubKey
}

func (pm *PoolAddressDummyMgr) SetObservedNextPoolAddrPubKey(ppks common.PoolPubKeys) {
	pm.observedNextPoolAddrPubKey = ppks
}

// BeginBlock should be called when BeginBlock
func (pm *PoolAddressDummyMgr) BeginBlock(_ sdk.Context) error {
	return kaboom
}

// EndBlock contains some actions THORNode need to take when block commit
func (pm *PoolAddressDummyMgr) EndBlock(_ct sdk.Context, _ *TxOutStore) {}

func (pm *PoolAddressDummyMgr) rotatePoolAddress(_ sdk.Context, _ *TxOutStore) {
}
