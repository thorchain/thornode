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
	addrs := NewPoolAddresses(GetRandomPoolPubKeys(), GetRandomPoolPubKeys(), GetRandomPoolPubKeys())
	return &PoolAddressDummyMgr{
		currentPoolAddresses: addrs,
	}
}

func (pm *PoolAddressDummyMgr) GetAsgardPoolPubKey(chain common.Chain) *common.PoolPubKey {
	return pm.GetCurrentPoolAddresses().Current.GetByChain(chain)
}

// GetCurrentPoolAddresses return current pool addresses
func (pm *PoolAddressDummyMgr) GetCurrentPoolAddresses() *PoolAddresses {
	return pm.currentPoolAddresses
}

func (pm *PoolAddressDummyMgr) SetRotateWindowOpen(b bool) {
	pm.isRotateWindowOpen = b
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
func (pm *PoolAddressDummyMgr) EndBlock(_ct sdk.Context, _ TxOutStore) {}

func (pm *PoolAddressDummyMgr) RotatePoolAddress(_ sdk.Context, _ common.PoolPubKeys, _ TxOutStore) {}

func (pm *PoolAddressDummyMgr) rotatePoolAddress(_ sdk.Context, _ TxOutStore) {}
