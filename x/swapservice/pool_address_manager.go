package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

// const values used to emit events
const (
	EventTypeNewPoolAddress = `pooladdress_new`
	EventTypePoolAddress    = `pooladdress`
	PoolAddressAction       = `action`
)

// PoolAddressManager is going to manage the pool addresses , rotate etc
type PoolAddressManager struct {
	k                          Keeper
	currentPoolAddresses       PoolAddresses
	ObservedNextPoolAddrPubKey common.PubKey
}

// NewPoolAddressManager create a new PoolAddressManager
func NewPoolAddressManager(k Keeper) *PoolAddressManager {
	return &PoolAddressManager{
		k: k,
	}
}

func (pm *PoolAddressManager) GetCurrentPoolAddresses() PoolAddresses {
	return pm.currentPoolAddresses
}

// BeginBlock
func (pm *PoolAddressManager) BeginBlock(ctx sdk.Context, height int64) {
	// decide pool addresses
	if pm.currentPoolAddresses.IsEmpty() {
		pm.currentPoolAddresses = pm.k.GetPoolAddresses(ctx)
	}
}

func (pm *PoolAddressManager) EndBlock(ctx sdk.Context, height int64, store *TxOutStore) {
	pm.currentPoolAddresses = pm.rotatePoolAddress(ctx, height, pm.currentPoolAddresses, store)
	pm.k.SetPoolAddresses(ctx, pm.currentPoolAddresses)
}

func (pm *PoolAddressManager) rotatePoolAddress(ctx sdk.Context, height int64, poolAddresses PoolAddresses, store *TxOutStore) PoolAddresses {
	if poolAddresses.IsEmpty() {
		ctx.Logger().Error("current pool addresses is nil , something is wrong")
	}
	// it is not time to rotate yet
	if poolAddresses.RotateAt > height {
		return poolAddresses
	}
	rotatePerBlockHeight := pm.k.GetAdminConfigRotatePerBlockHeight(ctx, sdk.AccAddress{})
	newPoolAddresses := NewPoolAddresses(poolAddresses.Current, poolAddresses.Next, common.NoAddress, height+rotatePerBlockHeight)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeNewPoolAddress,
			sdk.NewAttribute("current pool address", newPoolAddresses.Current.String()),
			sdk.NewAttribute("next pool address", newPoolAddresses.Next.String()),
			sdk.NewAttribute("previous pool address", newPoolAddresses.Previous.String())))
	if err := moveAssetsToNewPool(ctx, pm.k, store, newPoolAddresses); err != nil {
		ctx.Logger().Error("fail to move assets to new pool", err)
	}

	return newPoolAddresses
}

// move all assets based on pool balance to new pool
func moveAssetsToNewPool(ctx sdk.Context, k Keeper, store *TxOutStore, addresses PoolAddresses) error {
	// pool address actually didn't changed , so don't need to move asset
	if addresses.Previous.Equals(addresses.Current) {
		return nil
	}
	iter := k.GetPoolDataIterator(ctx)
	defer iter.Close()
	runeTotal := sdk.ZeroUint()
	poolRefundGas := k.GetAdminConfigInt64(ctx, PoolRefundGasKey, PoolRefundGasKey.Default(), sdk.AccAddress{})
	coins := common.Coins{}
	for ; iter.Valid(); iter.Next() {
		var p Pool
		err := k.cdc.UnmarshalBinaryBare(iter.Value(), &p)
		if err != nil {
			return errors.Wrap(err, "fail to unmarshal pool")
		}
		assetAmount := p.BalanceAsset
		// we only take BNB for now
		if common.IsBNBAsset(p.Asset) {
			assetAmount = assetAmount.Sub(sdk.NewUint(uint64(poolRefundGas)))
		}
		runeTotal = runeTotal.Add(p.BalanceRune)
		if p.BalanceAsset.GT(sdk.ZeroUint()) {
			coins = append(coins, common.NewCoin(p.Asset, assetAmount))
		}
	}

	allNodeAccounts, err := k.ListNodeAccounts(ctx)
	if nil != err {
		return errors.Wrap(err, "fail to get all node accounts")
	}

	// Validator bond paid to the pool as well , let's make sure all the bond get se
	for _, item := range allNodeAccounts {
		runeTotal = runeTotal.Add(item.Bond)
	}

	if !runeTotal.IsZero() {
		coins = append(coins, common.NewCoin(common.RuneA1FAsset, runeTotal))
	}
	if len(coins) > 0 {
		store.AddTxOutItem(ctx, k, &TxOutItem{
			PoolAddress: addresses.Previous,
			ToAddress:   addresses.Current,
			Coins:       coins,
		}, true)
	}
	return nil
}
