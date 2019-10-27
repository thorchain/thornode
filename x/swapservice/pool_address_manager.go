package swapservice

import (
	"fmt"

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
	IsRotateWindowOpen         bool
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

	if height >= pm.currentPoolAddresses.RotateWindowOpenAt && height < pm.currentPoolAddresses.RotateAt {
		pm.IsRotateWindowOpen = true
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
	windowOpen := pm.k.GetAdminConfigValidatorsChangeWindow(ctx, sdk.AccAddress{})
	rotateAt := height + rotatePerBlockHeight
	windowOpenAt := rotateAt - windowOpen
	newPoolAddresses := NewPoolAddresses(poolAddresses.Current, poolAddresses.Next, common.EmptyPubKey, rotateAt, windowOpenAt)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeNewPoolAddress,
			sdk.NewAttribute("current pool pub key", newPoolAddresses.Current.String()),
			sdk.NewAttribute("next pool pub key", newPoolAddresses.Next.String()),
			sdk.NewAttribute("previous pool pub key", newPoolAddresses.Previous.String())))
	if err := moveAssetsToNewPool(ctx, pm.k, store, newPoolAddresses); err != nil {
		ctx.Logger().Error("fail to move assets to new pool", err)
	}

	return newPoolAddresses
}

func getAllChains(ctx sdk.Context, k Keeper) ([]common.Chain, error) {
	chainMap := make(map[common.Chain]bool)
	chains := make([]common.Chain, 0)
	iter := k.GetPoolDataIterator(ctx)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var p Pool
		err := k.cdc.UnmarshalBinaryBare(iter.Value(), &p)
		if err != nil {
			return nil, errors.Wrap(err, "fail to unmarshal pool")
		}
		if chainMap[p.Asset.Chain] {
			continue
		}
		chains = append(chains, p.Asset.Chain)
		chainMap[p.Asset.Chain] = true
	}
	return chains, nil
}

// move all assets based on pool balance to new pool
func moveAssetsToNewPool(ctx sdk.Context, k Keeper, store *TxOutStore, addresses PoolAddresses) error {
	// pool address actually didn't changed , so don't need to move asset
	if addresses.Previous.Equals(addresses.Current) {
		return nil
	}
	chains, err := getAllChains(ctx, k)
	if nil != err {
		return fmt.Errorf("fail to get all chains from pool,err:%w", err)
	}
	runeTotal := sdk.ZeroUint()
	for _, c := range chains {
		runeAmount, err := moveChainAssetToNewPool(ctx, k, store, c, addresses)
		if nil != err {
			return fmt.Errorf("fail to move asset for chain %s,%w", c, err)
		}
		runeTotal = runeTotal.Add(runeAmount)
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
		fromAddr, err := addresses.Previous.GetAddress(common.BNBChain)
		if nil != err {
			return fmt.Errorf("fail to get address for chain %s from pub key %s ,err:%w", common.BNBChain, addresses.Previous, err)
		}
		toAddr, err := addresses.Current.GetAddress(common.BNBChain)
		if nil != err {
			return fmt.Errorf("fail to get address for chain %s from pub key %s ,err:%w", common.BNBChain, addresses.Current, err)
		}
		store.AddTxOutItem(ctx, k, &TxOutItem{
			PoolAddress: fromAddr,
			ToAddress:   toAddr,
			Coins: common.Coins{
				common.NewCoin(common.RuneA1FAsset, runeTotal),
			},
		}, true)
	}

	return nil
}

func moveChainAssetToNewPool(ctx sdk.Context, k Keeper, store *TxOutStore, chain common.Chain, addresses PoolAddresses) (sdk.Uint, error) {
	iter := k.GetPoolDataIterator(ctx)
	defer iter.Close()
	runeTotal := sdk.ZeroUint()
	poolRefundGas := k.GetAdminConfigInt64(ctx, PoolRefundGasKey, PoolRefundGasKey.Default(), sdk.AccAddress{})
	coins := common.Coins{}
	for ; iter.Valid(); iter.Next() {
		var p Pool
		err := k.cdc.UnmarshalBinaryBare(iter.Value(), &p)
		if err != nil {
			return sdk.ZeroUint(), errors.Wrap(err, "fail to unmarshal pool")
		}
		if !chain.Equals(p.Asset.Chain) {
			continue
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
	fromAddr, err := addresses.Previous.GetAddress(chain)
	if nil != err {
		return sdk.ZeroUint(), fmt.Errorf("fail to get address for chain %s from pub key %s ,err:%w", chain, addresses.Previous, err)
	}
	toAddr, err := addresses.Current.GetAddress(chain)
	if nil != err {
		return sdk.ZeroUint(), fmt.Errorf("fail to get address for chain %s from pub key %s ,err:%w", chain, addresses.Current, err)
	}
	if len(coins) > 0 {
		store.AddTxOutItem(ctx, k, &TxOutItem{
			PoolAddress: fromAddr,
			ToAddress:   toAddr,
			Coins:       coins,
		}, true)
	}
	return runeTotal, nil
}
