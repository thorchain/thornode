package types

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

// Yggdrasil Pool
type Yggdrasil struct {
	PubKey common.PubKey `json:"pub_key"`
	Coins  common.Coins  `json:"coins"`
}

type Yggdrasils []Yggdrasil

func NewYggdrasil(pk common.PubKey) Yggdrasil {
	return Yggdrasil{
		PubKey: pk,
		Coins:  make(common.Coins, 0),
	}
}

func (ygg Yggdrasil) IsEmpty() bool {
	return ygg.PubKey.IsEmpty()
}

// IsValid check whether Yggdrasil has all necessary values
func (ygg Yggdrasil) IsValid() error {
	if ygg.PubKey.IsEmpty() {
		return errors.New("pubkey cannot be empty")
	}
	return nil
}

func (ygg Yggdrasil) HasFunds() bool {
	amt := sdk.ZeroUint()
	for _, coin := range ygg.Coins {
		amt = amt.Add(coin.Amount)
	}
	return !amt.IsZero()
}

// Check if this yggdrasil has a particular asset
func (ygg Yggdrasil) HasAsset(asset common.Asset) bool {
	return !ygg.GetCoin(asset).Amount.IsZero()
}

func (ygg Yggdrasil) GetCoin(asset common.Asset) common.Coin {
	for _, coin := range ygg.Coins {
		if coin.Asset.Equals(asset) {
			return coin
		}
	}
	return common.NewCoin(asset, sdk.ZeroUint())
}

func (ygg *Yggdrasil) AddFunds(coins common.Coins) {
	for _, coin := range coins {
		if ygg.HasAsset(coin.Asset) {
			for i, ycoin := range ygg.Coins {
				if coin.Asset.Equals(ycoin.Asset) {
					ygg.Coins[i].Amount = ycoin.Amount.Add(coin.Amount)
				}
			}
		} else {
			ygg.Coins = append(ygg.Coins, coin)
		}
	}
}

func (ygg *Yggdrasil) SubFunds(coins common.Coins) {
	for _, coin := range coins {
		for i, ycoin := range ygg.Coins {
			if coin.Asset.Equals(ycoin.Asset) {
				// safeguard to protect against enter negative values
				if coin.Amount.GTE(ycoin.Amount) {
					coin.Amount = ycoin.Amount
				}
				ygg.Coins[i].Amount = ycoin.Amount.Sub(coin.Amount)
			}
		}
	}
}

func (yggs Yggdrasils) SortBy(sortBy common.Asset) Yggdrasils {
	// use the ygg pool with the highest quantity of our coin
	sort.Slice(yggs[:], func(i, j int) bool {
		return yggs[i].GetCoin(sortBy).Amount.GT(
			yggs[j].GetCoin(sortBy).Amount,
		)
	})

	return yggs
}
