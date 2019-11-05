package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

// const values used to emit events
const (
	EventTypeNewPoolAddress = `NewPoolAddress`
)

// PoolAddressManager is going to manage the pool addresses , rotate etc
type PoolAddressManager struct {
	k                          Keeper
	currentPoolAddresses       *PoolAddresses
	ObservedNextPoolAddrPubKey common.PoolPubKeys
	IsRotateWindowOpen         bool
}

// NewPoolAddressManager create a new PoolAddressManager
func NewPoolAddressManager(k Keeper) *PoolAddressManager {
	return &PoolAddressManager{
		k: k,
	}
}

// GetCurrentPoolAddresses return current pool addresses
func (pm *PoolAddressManager) GetCurrentPoolAddresses() *PoolAddresses {
	return pm.currentPoolAddresses
}

// BeginBlock should be called when BeginBlock
func (pm *PoolAddressManager) BeginBlock(ctx sdk.Context) {
	height := ctx.BlockHeight()
	// decide pool addresses
	if pm.currentPoolAddresses == nil || pm.currentPoolAddresses.IsEmpty() {
		poolAddresses := pm.k.GetPoolAddresses(ctx)
		pm.currentPoolAddresses = &poolAddresses
	}

	if height >= pm.currentPoolAddresses.RotateWindowOpenAt && height < pm.currentPoolAddresses.RotateAt {
		if pm.IsRotateWindowOpen {
			return
		}

		pm.IsRotateWindowOpen = true
	}
}

// EndBlock contains some actions we need to take when block commit
func (pm *PoolAddressManager) EndBlock(ctx sdk.Context, store *TxOutStore) {
	if nil == pm.currentPoolAddresses {
		return
	}
	pm.rotatePoolAddress(ctx, store)
	pm.k.SetPoolAddresses(ctx, pm.currentPoolAddresses)
}

func (pm *PoolAddressManager) rotatePoolAddress(ctx sdk.Context, store *TxOutStore) {
	poolAddresses := pm.currentPoolAddresses
	if ctx.BlockHeight() == 1 {
		// we don't need to do anything on
		return
	}
	if poolAddresses.IsEmpty() {
		ctx.Logger().Error("current pool addresses is nil , something is wrong")
		return
	}
	// likely there is a configuration error
	if poolAddresses.RotateAt == 0 {
		ctx.Logger().Error("rotate at block height had been set at 0, likely there is configuration error")
		return
	}

	height := ctx.BlockHeight()
	// it is not time to rotate yet
	if poolAddresses.RotateAt > height {
		return
	}

	rotatePerBlockHeight := pm.k.GetAdminConfigRotatePerBlockHeight(ctx, sdk.AccAddress{})
	windowOpen := pm.k.GetAdminConfigValidatorsChangeWindow(ctx, sdk.AccAddress{})
	rotateAt := height + rotatePerBlockHeight
	windowOpenAt := rotateAt - windowOpen
	pm.currentPoolAddresses = NewPoolAddresses(poolAddresses.Current, poolAddresses.Next, common.EmptyPoolPubKeys, rotateAt, windowOpenAt)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(EventTypeNewPoolAddress,
			sdk.NewAttribute("current pool pub key", pm.currentPoolAddresses.Current.String()),
			sdk.NewAttribute("next pool pub key", pm.currentPoolAddresses.Next.String()),
			sdk.NewAttribute("previous pool pub key", pm.currentPoolAddresses.Previous.String())))
	if err := moveAssetsToNewPool(ctx, pm.k, store, pm.currentPoolAddresses); err != nil {
		ctx.Logger().Error("fail to move assets to new pool", err)
	}
}

// move all assets based on pool balance to new pool
func moveAssetsToNewPool(ctx sdk.Context, k Keeper, store *TxOutStore, addresses *PoolAddresses) error {
	chains := k.GetChains(ctx)
	runeTotal := sdk.ZeroUint()
	for _, c := range chains {
		if c.Equals(common.BNBChain) {
			continue
		}
		runeAmount, err := moveChainAssetToNewPool(ctx, k, store, c, addresses)
		if nil != err {
			return fmt.Errorf("fail to move asset for chain %s,%w", c, err)
		}
		runeTotal = runeTotal.Add(runeAmount)
	}
	// we must have BNB chain
	return moveBNBChainAssetToNewPool(ctx, k, store, runeTotal, addresses)
}

func moveChainAssetToNewPool(ctx sdk.Context, k Keeper, store *TxOutStore, chain common.Chain, addresses *PoolAddresses) (sdk.Uint, error) {
	currentAddr := addresses.Current.GetByChain(chain)
	previousAddr := addresses.Previous.GetByChain(chain)
	if currentAddr.Equals(previousAddr) {
		// nothing to move
		return sdk.ZeroUint(), nil
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
			return runeTotal, errors.Wrap(err, "fail to unmarshal pool")
		}
		if !p.Asset.Chain.Equals(chain) {
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

	toAddr, err := currentAddr.PubKey.GetAddress(chain)
	if nil != err {
		return sdk.ZeroUint(), fmt.Errorf("fail to get address for chain %s from pub key %s ,err:%w", chain, addresses.Current, err)
	}
	if len(coins) > 0 {
		store.AddTxOutItem(ctx, k, &TxOutItem{
			Chain:       currentAddr.Chain,
			PoolAddress: currentAddr.PubKey,
			ToAddress:   toAddr,
			Coins:       coins,
		}, true)
	}
	return runeTotal, nil
}

func moveBNBChainAssetToNewPool(ctx sdk.Context, k Keeper, store *TxOutStore, runeTotal sdk.Uint, addresses *PoolAddresses) error {
	currentAddr := addresses.Current.GetByChain(common.BNBChain)
	previousAddr := addresses.Previous.GetByChain(common.BNBChain)
	if currentAddr.Equals(previousAddr) {
		// nothing to move
		return nil
	}
	iter := k.GetPoolDataIterator(ctx)
	defer iter.Close()

	poolRefundGas := k.GetAdminConfigInt64(ctx, PoolRefundGasKey, PoolRefundGasKey.Default(), sdk.AccAddress{})
	coins := common.Coins{}
	for ; iter.Valid(); iter.Next() {
		var p Pool
		err := k.cdc.UnmarshalBinaryBare(iter.Value(), &p)
		if err != nil {
			return errors.Wrap(err, "fail to unmarshal pool")
		}
		if !p.Asset.Chain.Equals(common.BNBChain) {
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
	allNodeAccounts, err := k.ListNodeAccounts(ctx)
	if nil != err {
		return errors.Wrap(err, "fail to get all node accounts")
	}

	// Validator bond paid to the pool as well , let's make sure all the bond get sent to new pool
	for _, item := range allNodeAccounts {
		runeTotal = runeTotal.Add(item.Bond)
	}

	if !runeTotal.IsZero() {
		coins = append(coins, common.NewCoin(common.RuneAsset(), runeTotal))
	}
	toAddr, err := currentAddr.PubKey.GetAddress(common.BNBChain)
	if nil != err {
		return fmt.Errorf("fail to get address for chain %s from pub key %s ,err:%w", common.BNBChain, addresses.Current, err)
	}
	if len(coins) > 0 {
		store.AddTxOutItem(ctx, k, &TxOutItem{
			Chain:       currentAddr.Chain,
			PoolAddress: currentAddr.PubKey,
			ToAddress:   toAddr,
			Coins:       coins,
		}, true)
	}
	return nil
}
