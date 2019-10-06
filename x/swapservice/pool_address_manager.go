package swapservice

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/common"
)

const rotatePoolAddressAfterBlocks int64 = 100

// PoolAddressManager is going to manage the pool addresses , rotate etc
type PoolAddressManager struct {
	k                  Keeper
	currentPoolAddress PoolAddresses
}

// NewPoolAddressManager create a new PoolAddressManager
func NewPoolAddressManager(k Keeper) *PoolAddressManager {
	return &PoolAddressManager{
		k: k,
	}
}

func (pm *PoolAddressManager) GetCurrentPoolAddresses() PoolAddresses {
	return pm.currentPoolAddress
}

// BeginBlock
func (pm *PoolAddressManager) BeginBlock(ctx sdk.Context, height int64) {
	// decide pool addresses
	if height == 1 {
		pa, err := pm.setupInitialPoolAddresses(ctx, height)
		if nil != err {
			ctx.Logger().Error("fail to setup initial pool address", err)
		}
		pm.currentPoolAddress = pa
	}
	if pm.currentPoolAddress.IsEmpty() {
		pm.currentPoolAddress = pm.k.GetPoolAddresses(ctx)
	}
}

func (pm *PoolAddressManager) EndBlock(ctx sdk.Context, height int64, store *TxOutStore) {

	pm.currentPoolAddress = pm.rotatePoolAddress(ctx, height, pm.currentPoolAddress, store)
	pm.k.SetPoolAddresses(ctx, pm.currentPoolAddress)
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
	newPoolAddresses := NewPoolAddresses(poolAddresses.Current, poolAddresses.Next, next.Accounts.SignerBNBAddress, height+rotatePoolAddressAfterBlocks)
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
	if totalActiveAccounts == 1 {
		na := nodeAccounts[0]
		ctx.Logger().Info("only one active node account, no pool rotation")
		return NewPoolAddresses(common.NoBnbAddress, na.Accounts.SignerBNBAddress, na.Accounts.SignerBNBAddress, height+rotatePoolAddressAfterBlocks), nil

	}
	sort.Sort(nodeAccounts)
	na := nodeAccounts[0]
	sec := nodeAccounts[1]
	ctx.Logger().Info("two or more active nodes , we will rotate pools")
	return NewPoolAddresses(common.NoBnbAddress, na.Accounts.SignerBNBAddress, sec.Accounts.SignerBNBAddress, height+rotatePoolAddressAfterBlocks), nil

}
