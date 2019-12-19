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
	TriggerKeygen(ctx sdk.Context, nas NodeAccounts) error
}

// VaultMgr is going to manage the vaults
type VaultMgr struct {
	k Keeper
}

// NewVaultMgr create a new vault manager
func NewVaultMgr(k Keeper) *VaultMgr {
	return &VaultMgr{k: k}
}

func (vm *VaultMgr) TriggerKeygen(ctx sdk.Context, nas NodeAccounts) error {
	keygen := make(Keygen, len(nas))
	for i := range nas {
		keygen[i] = nas[i].NodePubKey.Secp256k1
	}
	keygens := NewKeygens(uint64(ctx.BlockHeight()))
	keygens.Keygens = []Keygen{keygen}
	return vm.k.SetKeygens(ctx, keygens)
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
