package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/thornode/common"
)

// const values used to emit events
const (
	EventTypeNewPoolAddress    = `NewPoolAddress`
	EventTypeAbortPoolRotation = "AbortPoolRotation"
)

type PoolAddressManager interface {
	BeginBlock(_ sdk.Context) error
	RotatePoolAddress(_ sdk.Context, _ common.PoolPubKeys, _ TxOutStore)
	GetCurrentPoolAddresses() *PoolAddresses
	GetAsgardPoolPubKey(_ common.Chain) *common.PoolPubKey
	SetObservedNextPoolAddrPubKey(ppks common.PoolPubKeys)
	ObservedNextPoolAddrPubKey() common.PoolPubKeys
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

func (pm *PoolAddressMgr) GetAsgardPoolPubKey(chain common.Chain) *common.PoolPubKey {
	return pm.GetCurrentPoolAddresses().Current.GetByChain(chain)
}

// BeginBlock should be called when BeginBlock
func (pm *PoolAddressMgr) BeginBlock(ctx sdk.Context) error {
	if pm.currentPoolAddresses == nil || pm.currentPoolAddresses.IsEmpty() {
		poolAddresses, err := pm.k.GetPoolAddresses(ctx)
		if err != nil {
			return err
		}
		pm.currentPoolAddresses = &poolAddresses
	}

	return nil
}

func (pm *PoolAddressMgr) RotatePoolAddress(ctx sdk.Context, poolpubkeys common.PoolPubKeys, store TxOutStore) {
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
		store.AddTxOutItem(ctx, &TxOutItem{
			Chain:       currentAddr.Chain,
			VaultPubKey: previousAddr.PubKey,
			InHash:      common.BlankTxID,
			ToAddress:   toAddr,
			Coin:        coin,
		})
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
		store.AddTxOutItem(ctx, &TxOutItem{
			Chain:       currentAddr.Chain,
			VaultPubKey: previousAddr.PubKey,
			InHash:      common.BlankTxID,
			ToAddress:   toAddr,
			Coin:        coin,
		})
	}
	return nil
}
