package swapservice

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/common"
)

// const values used to emit events
const (
	EventTypeNewPoolAddress = `pooladdress_new`
	EventTypePoolAddress    = `pooladdress`
	PoolAddressAction       = `action`
)

// PoolAddressManager is going to manage the pool addresses , rotate etc
type PoolAddressManager struct {
	k                    Keeper
	currentPoolAddresses PoolAddresses
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
	if height == 1 {
		pa, err := pm.setupInitialPoolAddresses(ctx, height)
		if nil != err {
			ctx.Logger().Error("fail to setup initial pool address", err)
		}
		pm.currentPoolAddresses = pa
	}
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
	nodeAccounts, err := pm.k.ListActiveNodeAccounts(ctx)
	if nil != err {
		ctx.Logger().Error("fail to get active node accounts", "err", err)
		return poolAddresses
	}
	sort.Sort(nodeAccounts)
	next := nodeAccounts.After(poolAddresses.Next)
	rotatePerBlockHeight := pm.k.GetAdminConfigRotatePerBlockHeight(ctx, sdk.AccAddress{})
	newPoolAddresses := NewPoolAddresses(poolAddresses.Current, poolAddresses.Next, next.Accounts.SignerBNBAddress, height+rotatePerBlockHeight)
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

	for ; iter.Valid(); iter.Next() {
		var p Pool
		err := k.cdc.UnmarshalBinaryBare(iter.Value(), &p)
		if err != nil {
			return errors.Wrap(err, "fail to unmarshal pool")
		}
		runeTotal = runeTotal.Add(p.BalanceRune)
		if p.BalanceToken.GT(sdk.ZeroUint()) {
			store.AddTxOutItem(&TxOutItem{
				PoolAddress: addresses.Previous,
				ToAddress:   addresses.Current,
				Coins: common.Coins{
					common.NewCoin(p.Ticker, p.BalanceToken),
				},
			})
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
		store.AddTxOutItem(&TxOutItem{
			PoolAddress: addresses.Previous,
			ToAddress:   addresses.Current,
			Coins: common.Coins{
				common.NewCoin(common.RuneTicker, runeTotal),
			},
		})
	}
	return nil
}

var emptyPoolAddresses PoolAddresses

func (pm *PoolAddressManager) setupInitialPoolAddresses(ctx sdk.Context, height int64) (PoolAddresses, error) {
	// this method will only take effect when statechain started
	if height != 1 {
		return emptyPoolAddresses, errors.New("only setup initial pool address when chain start")
	}
	ctx.Logger().Info("setup initial pool addresses")
	nodeAccounts, err := pm.k.ListActiveNodeAccounts(ctx)
	if nil != err {
		ctx.Logger().Error("fail to get active node accounts", "err", err)
		return emptyPoolAddresses, errors.Wrap(err, "fail to get active node accounts")
	}
	totalActiveAccounts := len(nodeAccounts)
	if totalActiveAccounts == 0 {
		ctx.Logger().Error("no active node account")

		return emptyPoolAddresses, errors.New("no active node account")
	}
	rotatePerBlockHeight := pm.k.GetAdminConfigRotatePerBlockHeight(ctx, sdk.AccAddress{})
	if totalActiveAccounts == 1 {
		na := nodeAccounts[0]
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypePoolAddress,
				sdk.NewAttribute(PoolAddressAction, "no pool rotation"),
				sdk.NewAttribute("reason", "no active node account")))
		ctx.Logger().Info("only one active node account, no pool rotation")
		return NewPoolAddresses(common.NoBnbAddress, na.Accounts.SignerBNBAddress, na.Accounts.SignerBNBAddress, height+rotatePerBlockHeight), nil

	}
	sort.Sort(nodeAccounts)
	na := nodeAccounts[0]
	sec := nodeAccounts[1]
	ctx.Logger().Info("two or more active nodes , we will rotate pools")
	return NewPoolAddresses(common.NoBnbAddress, na.Accounts.SignerBNBAddress, sec.Accounts.SignerBNBAddress, height+rotatePerBlockHeight), nil

}
