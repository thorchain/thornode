package types

import (
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// VaultType there are two different types of Vault in thorchain
type VaultType string

const (
	UnknownVault   VaultType = "unknown"
	AsgardVault    VaultType = "asgard"
	YggdrasilVault VaultType = "yggdrasil"
)

type VaultStatus string

const (
	// ActiveVault means the vault is currently actively in use
	ActiveVault VaultStatus = "active"
	// RetiringVault means the vault is in the process of retiring
	RetiringVault VaultStatus = "retiring"
	// InactiveVault means the vault is not active anymore
	InactiveVault VaultStatus = "inactive"
)

// Vault usually represent the pool we are using
type Vault struct {
	BlockHeight           int64          `json:"block_height"`
	PubKey                common.PubKey  `json:"pub_key"`
	Coins                 common.Coins   `json:"coins"`
	Type                  VaultType      `json:"type"`
	Status                VaultStatus    `json:"status"`
	StatusSince           int64          `json:"status_since"`
	Membership            common.PubKeys `json:"membership"`
	Chains                common.Chains  `json:"chains"`
	InboundTxCount        int64          `json:"inbound_tx_count"`
	OutboundTxCount       int64          `json:"outbound_tx_count"`
	PendingTxBlockHeights []int64        `json:"pending_tx_heights"`
}

type Vaults []Vault

// NewVault create a new instance of vault
func NewVault(height int64, status VaultStatus, vtype VaultType, pk common.PubKey, chains common.Chains) Vault {
	return Vault{
		BlockHeight: height,
		StatusSince: height,
		PubKey:      pk,
		Coins:       make(common.Coins, 0),
		Type:        vtype,
		Status:      status,
		Chains:      chains,
	}
}

// IsType determine whether the vault is given type
func (v Vault) IsType(vtype VaultType) bool {
	return v.Type == vtype
}

// IsAsgard check whether the vault is Asgard vault, it returns true when it is an asgard vault
func (v Vault) IsAsgard() bool {
	return v.IsType(AsgardVault)
}

// IsYggdrasil return true when the vault is YggdrasilVault
func (v Vault) IsYggdrasil() bool {
	return v.IsType(YggdrasilVault)
}

// IsEmpty returns true when the vault pubkey is empty
func (v Vault) IsEmpty() bool {
	return v.PubKey.IsEmpty()
}

// Contains check whether the given pubkey is party of the originally node who create this vault
func (v Vault) Contains(pubkey common.PubKey) bool {
	return v.Membership.Contains(pubkey)
}

// UpdateStatus set the vault to given status
func (v *Vault) UpdateStatus(s VaultStatus, height int64) {
	v.Status = s
	v.StatusSince = height
}

// IsValid check whether Vault has all necessary values
func (v Vault) IsValid() error {
	if v.PubKey.IsEmpty() {
		return errors.New("pubkey cannot be empty")
	}
	return nil
}

// HasFunds check whether the vault pool has fund
func (v Vault) HasFunds() bool {
	for _, coin := range v.Coins {
		if !coin.Amount.IsZero() {
			return true
		}
	}
	return false
}

// HasFundsForChain check whether the vault pool has funds for a specific chain
func (v Vault) HasFundsForChain(chain common.Chain) bool {
	for _, coin := range v.Coins {
		if coin.Asset.Chain.Equals(chain) && !coin.Amount.IsZero() {
			return true
		}
	}
	return false
}

// CoinLength - counts the number of coins this vault has
func (v Vault) CoinLength() (count int) {
	for _, coin := range v.Coins {
		if !coin.Amount.IsZero() {
			count += 1
		}
	}
	return
}

// HasAsset Check if this vault has a particular asset
func (v Vault) HasAsset(asset common.Asset) bool {
	return !v.GetCoin(asset).Amount.IsZero()
}

// GetCoin return coin type of given asset
func (v Vault) GetCoin(asset common.Asset) common.Coin {
	for _, coin := range v.Coins {
		if coin.Asset.Equals(asset) {
			return coin
		}
	}
	return common.NewCoin(asset, sdk.ZeroUint())
}

// GetMembers return members who's address exist in the given list
func (v Vault) GetMembers(activeObservers []sdk.AccAddress) (common.PubKeys, error) {
	signers := common.PubKeys{}
	for _, k := range v.Membership {
		addr, err := k.GetThorAddress()
		if err != nil {
			return common.PubKeys{}, fmt.Errorf("fail to get thor address: %w", err)
		}
		for _, item := range activeObservers {
			if item.Equals(addr) {
				signers = append(signers, k)
			}
		}
	}
	return signers, nil
}

// AddFunds add given coins into vault
func (v *Vault) AddFunds(coins common.Coins) {
	for _, coin := range coins {
		found := false
		for i, ycoin := range v.Coins {
			if coin.Asset.Equals(ycoin.Asset) {
				v.Coins[i].Amount = ycoin.Amount.Add(coin.Amount)
				found = true
				break
			}
		}
		if found {
			continue
		}
		if !v.Chains.Has(coin.Asset.Chain) {
			v.Chains = append(v.Chains, coin.Asset.Chain)
		}
		v.Coins = append(v.Coins, coin)
	}
}

// SubFunds subtract given coins from vault
func (v *Vault) SubFunds(coins common.Coins) {
	for _, coin := range coins {
		for i, ycoin := range v.Coins {
			if coin.Asset.Equals(ycoin.Asset) {
				// safeguard to protect against enter negative values
				if coin.Amount.GTE(ycoin.Amount) {
					coin.Amount = ycoin.Amount
				}
				v.Coins[i].Amount = common.SafeSub(ycoin.Amount, coin.Amount)
			}
		}
	}
}

// AppendPendingTxBlockHeights will add current block height into the list , also remove the block height that is too old
func (v *Vault) AppendPendingTxBlockHeights(blockHeight int64, constAccessor constants.ConstantValues) {
	heights := []int64{
		blockHeight,
	}
	for _, item := range v.PendingTxBlockHeights {
		if (blockHeight - item) <= constAccessor.GetInt64Value(constants.SigningTransactionPeriod) {
			heights = append(heights, item)
		}
	}
	v.PendingTxBlockHeights = heights
}

// RemovePendingTxBlockHeights remove the given block height from internal pending tx block height
func (v *Vault) RemovePendingTxBlockHeights(blockHeight int64) {
	idxToRemove := -1
	for idx, item := range v.PendingTxBlockHeights {
		if item == blockHeight {
			idxToRemove = idx
			break
		}
	}
	if idxToRemove != -1 {
		v.PendingTxBlockHeights = append(v.PendingTxBlockHeights[:idxToRemove], v.PendingTxBlockHeights[idxToRemove+1:]...)
	}
}

// LenPendingTxBlockHeights count how many outstanding block heights in the vault
// if the a block height is older than SigningTransactionPeriod , it will ignore
func (v *Vault) LenPendingTxBlockHeights(currentBlockHeight int64, constAccessor constants.ConstantValues) int {
	total := 0
	for _, item := range v.PendingTxBlockHeights {
		if (currentBlockHeight - item) <= constAccessor.GetInt64Value(constants.SigningTransactionPeriod) {
			total++
		}
	}
	return total
}

// SortBy order coins by the given asset
func (vs Vaults) SortBy(sortBy common.Asset) Vaults {
	// use the vault pool with the highest quantity of our coin
	sort.SliceStable(vs[:], func(i, j int) bool {
		return vs[i].GetCoin(sortBy).Amount.GT(
			vs[j].GetCoin(sortBy).Amount,
		)
	})

	return vs
}

// SelectByMinCoin return the vault that has least of given asset
func (vs Vaults) SelectByMinCoin(asset common.Asset) (vault Vault) {
	for _, v := range vs {
		if vault.IsEmpty() || v.GetCoin(asset).Amount.LT(vault.GetCoin(asset).Amount) {
			vault = v
		}
	}

	return
}

// SelectByMaxCoin return the vault that has most of given asset
func (vs Vaults) SelectByMaxCoin(asset common.Asset) (vault Vault) {
	for _, v := range vs {
		if v.GetCoin(asset).Amount.GT(vault.GetCoin(asset).Amount) {
			vault = v
		}
	}

	return
}

// HasAddress will go through the vaults to determinate whether any of the vault match the given address on the given chain
func (vs Vaults) HasAddress(chain common.Chain, address common.Address) (bool, error) {
	for _, item := range vs {
		addr, err := item.PubKey.GetAddress(chain)
		if err != nil {
			return false, fmt.Errorf("fail to get address from (%s) for chain(%s)", item.PubKey, chain)
		}
		if addr.Equals(address) {
			return true, nil
		}
	}
	return false, nil
}
