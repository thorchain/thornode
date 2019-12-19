package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
)

// const values used to emit events
const (
	EventTypeNewVault = "NewVault"
)

type VaultManager interface {
	RotatePoolAddress(_ sdk.Context, _ common.PoolPubKeys, _ TxOutStore)
}

// VaultMgr is going to manage the vaults
type VaultMgr struct {
	k Keeper
}

// NewVaultMgr create a new vault manager
func NewVaultMgr(k Keeper, valdatorMgr ValidatorManager) *VaultMgr {
	return &VaultMgr{k: k}
}

func (vm *VaultMgr) RotatePoolAddress(ctx sdk.Context, poolpubkeys common.PoolPubKeys, store TxOutStore) {
	/*
		poolAddresses := pm.currentPoolAddresses
		pm.currentPoolAddresses = NewPoolAddresses(poolAddresses.Current, poolpubkeys, common.EmptyPoolPubKeys)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeNewPoolAddress,
				sdk.NewAttribute("current pool pub key", pm.currentPoolAddresses.Current.String()),
				sdk.NewAttribute("next pool pub key", pm.currentPoolAddresses.Next.String()),
				sdk.NewAttribute("previous pool pub key", pm.currentPoolAddresses.Previous.String())))
		if err := moveAssetsToNewPool(ctx, pm.k, store, pm.currentPoolAddresses); err != nil {
			ctx.Logger().Error("fail to move assets to new pool", err)
		}
	*/
}
