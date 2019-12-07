package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// const values used to emit events
const (
	EventTypeNewPoolAddress    = `NewPoolAddress`
	EventTypeAbortPoolRotation = "AbortPoolRotation"
)

type PoolAddressManager interface {
	GetCurrentPoolAddresses() *PoolAddresses
	BeginBlock(ctx sdk.Context) error
	EndBlock(ctx sdk.Context, store TxOutStore)
	SetObservedNextPoolAddrPubKey(ppks common.PoolPubKeys)
	ObservedNextPoolAddrPubKey() common.PoolPubKeys
	IsRotateWindowOpen() bool
	SetRotateWindowOpen(_ bool)
	rotatePoolAddress(ctx sdk.Context, store TxOutStore)
}

// PoolAddressMgr is going to manage the pool addresses , rotate etc
type PoolAddressMgr struct {
	k                          Keeper
	currentPoolAddresses       *PoolAddresses
	observedNextPoolAddrPubKey common.PoolPubKeys
	isRotateWindowOpen         bool
}

// NewPoolAddressMgr create a new PoolAddressMgr
func NewPoolAddressMgr(k Keeper) *PoolAddressMgr {
	return &PoolAddressMgr{
		k: k,
	}
}

// GetCurrentPoolAddresses return current pool addresses
func (pm *PoolAddressMgr) GetCurrentPoolAddresses() *PoolAddresses {
	return pm.currentPoolAddresses
}

func (pm *PoolAddressMgr) IsRotateWindowOpen() bool {
	return pm.isRotateWindowOpen
}

func (pm *PoolAddressMgr) SetRotateWindowOpen(b bool) {
	pm.isRotateWindowOpen = b
}

func (pm *PoolAddressMgr) ObservedNextPoolAddrPubKey() common.PoolPubKeys {
	return pm.observedNextPoolAddrPubKey
}

func (pm *PoolAddressMgr) SetObservedNextPoolAddrPubKey(ppks common.PoolPubKeys) {
	pm.observedNextPoolAddrPubKey = ppks
}

// BeginBlock should be called when BeginBlock
func (pm *PoolAddressMgr) BeginBlock(ctx sdk.Context) error {
	height := ctx.BlockHeight()
	// decide pool addresses
	if pm.currentPoolAddresses == nil || pm.currentPoolAddresses.IsEmpty() {
		poolAddresses, err := pm.k.GetPoolAddresses(ctx)
		if err != nil {
			return err
		}
		pm.currentPoolAddresses = &poolAddresses
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
func (pm *PoolAddressMgr) EndBlock(ctx sdk.Context, store TxOutStore) {
	if nil == pm.currentPoolAddresses {
		return
	}
	// pool rotation window open
	if pm.isRotateWindowOpen && ctx.BlockHeight() == pm.currentPoolAddresses.RotateWindowOpenAt {
		// instruct signer to kick off tss keygen ceremony
		store.AddTxOutItem(ctx, pm.k, &TxOutItem{
			Chain: common.BNBChain,
			// Leave ToAddress empty on purpose, signer will observe this txout, and then kick of tss keygen ceremony
			ToAddress:   "",
			PoolAddress: pm.currentPoolAddresses.Current.GetByChain(common.BNBChain).PubKey,
			Coin:        common.NewCoin(common.BNBAsset, sdk.NewUint(37501)),
			Memo:        "nextpool",
		}, true)
	}
	pm.rotatePoolAddress(ctx, store)
	pm.k.SetPoolAddresses(ctx, pm.currentPoolAddresses)
}

func (pm *PoolAddressMgr) rotatePoolAddress(ctx sdk.Context, store TxOutStore) {
	poolAddresses := pm.currentPoolAddresses
	if ctx.BlockHeight() == 1 {
		// THORNode don't need to do anything on
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

	if poolAddresses.Next.IsEmpty() {
		ctx.Logger().Error("next pool address has not been confirmed , abort pool rotation")
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeAbortPoolRotation, sdk.NewAttribute("reason", "no next pool address")))
		return
	}

	rotatePerBlockHeight := constants.RotatePerBlockHeight
	windowOpen := constants.ValidatorsChangeWindow
	rotateAt := height + int64(rotatePerBlockHeight)
	windowOpenAt := rotateAt - int64(windowOpen)
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
func moveAssetsToNewPool(ctx sdk.Context, k Keeper, store TxOutStore, addresses *PoolAddresses) error {
	chains, err := k.GetChains(ctx)
	if err != nil {
		return err
	}
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
	// THORNode must have BNB chain
	return moveBNBChainAssetToNewPool(ctx, k, store, runeTotal, addresses)
}

func moveChainAssetToNewPool(ctx sdk.Context, k Keeper, store TxOutStore, chain common.Chain, addresses *PoolAddresses) (sdk.Uint, error) {
	currentAddr := addresses.Current.GetByChain(chain)
	previousAddr := addresses.Previous.GetByChain(chain)
	if currentAddr.Equals(previousAddr) {
		// nothing to move
		return sdk.ZeroUint(), nil
	}
	iter := k.GetPoolIterator(ctx)
	defer iter.Close()
	runeTotal := sdk.ZeroUint()
	poolRefundGas := k.GetAdminConfigInt64(ctx, PoolRefundGasKey, PoolRefundGasKey.Default(), sdk.AccAddress{})
	coins := common.Coins{}
	for ; iter.Valid(); iter.Next() {
		var p Pool
		err := k.Cdc().UnmarshalBinaryBare(iter.Value(), &p)
		if err != nil {
			return runeTotal, errors.Wrap(err, "fail to unmarshal pool")
		}
		if !p.Asset.Chain.Equals(chain) {
			continue
		}
		assetAmount := p.BalanceAsset
		// THORNode only take BNB for now
		if p.Asset.IsBNB() {
			assetAmount = common.SafeSub(assetAmount, sdk.NewUint(uint64(poolRefundGas)))
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
	for _, coin := range coins {
		store.AddTxOutItem(ctx, k, &TxOutItem{
			Chain:       currentAddr.Chain,
			PoolAddress: previousAddr.PubKey,
			InHash:      common.BlankTxID,
			ToAddress:   toAddr,
			Coin:        coin,
		}, true)
	}
	return runeTotal, nil
}

func moveBNBChainAssetToNewPool(ctx sdk.Context, k Keeper, store TxOutStore, runeTotal sdk.Uint, addresses *PoolAddresses) error {
	currentAddr := addresses.Current.GetByChain(common.BNBChain)
	previousAddr := addresses.Previous.GetByChain(common.BNBChain)
	if currentAddr.Equals(previousAddr) {
		// nothing to move
		return nil
	}
	iter := k.GetPoolIterator(ctx)
	defer iter.Close()

	poolRefundGas := k.GetAdminConfigInt64(ctx, PoolRefundGasKey, PoolRefundGasKey.Default(), sdk.AccAddress{})
	coins := common.Coins{}
	for ; iter.Valid(); iter.Next() {
		var p Pool
		err := k.Cdc().UnmarshalBinaryBare(iter.Value(), &p)
		if err != nil {
			return errors.Wrap(err, "fail to unmarshal pool")
		}
		if !p.Asset.Chain.Equals(common.BNBChain) {
			continue
		}
		assetAmount := p.BalanceAsset
		// THORNode only take BNB for now
		if p.Asset.IsBNB() {
			assetAmount = common.SafeSub(assetAmount, sdk.NewUint(uint64(poolRefundGas)))
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
	for _, coin := range coins {
		store.AddTxOutItem(ctx, k, &TxOutItem{
			Chain:       currentAddr.Chain,
			PoolAddress: previousAddr.PubKey,
			InHash:      common.BlankTxID,
			ToAddress:   toAddr,
			Coin:        coin,
		}, true)
	}
	return nil
}
