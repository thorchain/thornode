package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
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
func (pm *PoolAddressDummyMgr) BeginBlock(ctx sdk.Context) error {
	height := ctx.BlockHeight()
	// decide pool addresses
	if pm.currentPoolAddresses == nil || pm.currentPoolAddresses.IsEmpty() {
		// do nothing
	}

	if height >= pm.currentPoolAddresses.RotateWindowOpenAt && height < pm.currentPoolAddresses.RotateAt {
		if pm.IsRotateWindowOpen() {
			return nil
		}
		pm.isRotateWindowOpen = true
	}
	return nil
}

// EndBlock contains some actions THORNode need to take when block commit
func (pm *PoolAddressDummyMgr) EndBlock(ctx sdk.Context, store *TxOutStore) {}

func (pm *PoolAddressDummyMgr) rotatePoolAddress(ctx sdk.Context, store *TxOutStore) {
	poolAddresses := pm.currentPoolAddresses
	if ctx.BlockHeight() == 1 {
		// THORNode don't need to do anything on
		return
	}
	if poolAddresses.IsEmpty() {
		return
	}
	// likely there is a configuration error
	if poolAddresses.RotateAt == 0 {
		return
	}

	height := ctx.BlockHeight()
	// it is not time to rotate yet
	if poolAddresses.RotateAt > height {
		return
	}

	if poolAddresses.Next.IsEmpty() {
		return
	}

	rotatePerBlockHeight := constants.RotatePerBlockHeight
	windowOpen := constants.ValidatorsChangeWindow
	rotateAt := height + int64(rotatePerBlockHeight)
	windowOpenAt := rotateAt - int64(windowOpen)
	pm.currentPoolAddresses = NewPoolAddresses(poolAddresses.Current, poolAddresses.Next, common.EmptyPoolPubKeys, rotateAt, windowOpenAt)
}
