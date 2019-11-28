package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"gitlab.com/thorchain/thornode/common"
)

type KeeperNodeAccount interface {
	TotalNodeAccounts(ctx sdk.Context) (count int)
	TotalActiveNodeAccount(ctx sdk.Context) (int, error)
	ListNodeAccounts(ctx sdk.Context) (NodeAccounts, error)
	ListNodeAccountsByStatus(ctx sdk.Context, status NodeStatus) (NodeAccounts, error)
	ListActiveNodeAccounts(ctx sdk.Context) (NodeAccounts, error)
	GetLowestActiveVersion(ctx sdk.Context) int64
	IsWhitelistedNode(ctx sdk.Context, addr sdk.AccAddress) bool
	GetNodeAccount(ctx sdk.Context, addr sdk.AccAddress) (NodeAccount, error)
	GetNodeAccountByPubKey(ctx sdk.Context, pk common.PubKey) (NodeAccount, error)
	GetNodeAccountByBondAddress(ctx sdk.Context, addr common.Address) (NodeAccount, error)
	SetNodeAccount(ctx sdk.Context, na NodeAccount)
	SlashNodeAccountBond(ctx sdk.Context, na *NodeAccount, slash sdk.Uint)
	SlashNodeAccountRewards(ctx sdk.Context, na *NodeAccount, pts int64)
	EnsureTrustAccountUnique(ctx sdk.Context, consensusPubKey string, pubKeys common.PubKeys) error
	GetNodeAccountIterator(ctx sdk.Context) sdk.Iterator
}

// TotalNodeAccounts counts the number of trust accounts
func (k KVStore) TotalNodeAccounts(ctx sdk.Context) (count int) {
	nodes, _ := k.ListActiveNodeAccounts(ctx)
	return len(nodes)
}

// TotalActiveNodeAccount count the number of active node account
func (k KVStore) TotalActiveNodeAccount(ctx sdk.Context) (int, error) {
	activeNodes, err := k.ListActiveNodeAccounts(ctx)
	return len(activeNodes), err
}

// ListNodeAccounts - gets a list of all trust accounts
func (k KVStore) ListNodeAccounts(ctx sdk.Context) (NodeAccounts, error) {
	nodeAccounts := make(NodeAccounts, 0)
	naIterator := k.GetNodeAccountIterator(ctx)
	defer naIterator.Close()
	for ; naIterator.Valid(); naIterator.Next() {
		var na NodeAccount
		if err := k.cdc.UnmarshalBinaryBare(naIterator.Value(), &na); nil != err {
			return nil, errors.Wrap(err, "fail to unmarshal node account")
		}
		nodeAccounts = append(nodeAccounts, na)
	}
	return nodeAccounts, nil
}

// ListNodeAccountsByStatus - get a list of node accounts with the given status
// if status = NodeUnknown, then it return everything
func (k KVStore) ListNodeAccountsByStatus(ctx sdk.Context, status NodeStatus) (NodeAccounts, error) {
	nodeAccounts := make(NodeAccounts, 0)
	allNodeAccounts, err := k.ListNodeAccounts(ctx)
	if nil != err {
		return nodeAccounts, fmt.Errorf("fail to get all node accounts, %w", err)
	}
	for _, item := range allNodeAccounts {
		if item.Status == status {
			nodeAccounts = append(nodeAccounts, item)
		}
	}
	return nodeAccounts, nil
}

// ListActiveNodeAccounts - get a list of active trust accounts
func (k KVStore) ListActiveNodeAccounts(ctx sdk.Context) (NodeAccounts, error) {
	return k.ListNodeAccountsByStatus(ctx, NodeActive)
}

// GetLowestActiveVersion - get version number of lowest active node
func (k KVStore) GetLowestActiveVersion(ctx sdk.Context) int64 {
	nodes, _ := k.ListActiveNodeAccounts(ctx)
	if len(nodes) > 0 {
		version := nodes[0].Version
		for _, na := range nodes {
			if na.Version < version {
				version = na.Version
			}
		}
		return version
	}
	return 0
}

// IsWhitelistedAccount check whether the given account is white listed
func (k KVStore) IsWhitelistedNode(ctx sdk.Context, addr sdk.AccAddress) bool {
	ctx.Logger().Debug("IsWhitelistedAccount", "account address", addr.String())
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixNodeAccount, addr.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixNodeAccount))
	return store.Has([]byte(key))
}

// GetNodeAccount try to get node account with the given address from db
func (k KVStore) GetNodeAccount(ctx sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	ctx.Logger().Debug("GetNodeAccount", "node account", addr.String())
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixNodeAccount, addr.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixNodeAccount))
	payload := store.Get([]byte(key))
	var na NodeAccount
	if err := k.cdc.UnmarshalBinaryBare(payload, &na); nil != err {
		return na, errors.Wrap(err, "fail to unmarshal node account")
	}
	return na, nil
}

// GetNodeAccountByPubKey try to get node account with the given pubkey from db
func (k KVStore) GetNodeAccountByPubKey(ctx sdk.Context, pk common.PubKey) (NodeAccount, error) {
	addr, err := pk.GetThorAddress()
	if err != nil {
		return NodeAccount{}, err
	}
	return k.GetNodeAccount(ctx, addr)
}

// GetNodeAccountByBondAddress go through data store to get node account by it's signer bnb address
func (k KVStore) GetNodeAccountByBondAddress(ctx sdk.Context, addr common.Address) (NodeAccount, error) {
	ctx.Logger().Debug("GetNodeAccountByBondAddress", "signer bnb address", addr.String())
	var na NodeAccount
	nodeAccounts, err := k.ListNodeAccounts(ctx)
	if nil != err {
		return na, fmt.Errorf("fail to get all node accounts, %w", err)
	}
	for _, item := range nodeAccounts {
		if item.BondAddress.Equals(addr) {
			return item, nil
		}
	}
	return na, nil
}

// SetNodeAccount save the given node account into datastore
func (k KVStore) SetNodeAccount(ctx sdk.Context, na NodeAccount) {
	ctx.Logger().Debug("SetNodeAccount", "node account", na.String())
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixNodeAccount, na.NodeAddress.String(), getVersion(k.GetLowestActiveVersion(ctx), prefixNodeAccount))
	if na.Status == NodeActive {
		if na.ActiveBlockHeight == 0 {
			// the na is active, and does not have a block height when they
			// became active. This must be the first block they are active, so
			// we will set it now.
			na.ActiveBlockHeight = ctx.BlockHeight()
			na.SlashPoints = 0 // reset slash points
		}
	} else {
		if na.ActiveBlockHeight > 0 {
			// The na seems to have become a non active na. Therefore, lets
			// give them their bond rewards.
			vault := k.GetVaultData(ctx)
			// Find number of blocks they have been active for
			blockCount := ctx.BlockHeight() - na.ActiveBlockHeight
			blocks := calculateNodeAccountBondUints(ctx.BlockHeight(), na.ActiveBlockHeight, na.SlashPoints)
			// calc number of rune they are awarded
			reward := calcNodeRewards(blocks, vault.TotalBondUnits, vault.BondRewardRune)
			if reward.GT(sdk.ZeroUint()) {
				na.Bond = na.Bond.Add(reward)
				if vault.TotalBondUnits.GTE(sdk.NewUint(uint64(blockCount))) {
					vault.TotalBondUnits = vault.TotalBondUnits.Sub(sdk.NewUint(uint64(blockCount)))
				} else {
					vault.TotalBondUnits = sdk.ZeroUint()
				}
				// Minus the number of units na has (do not include slash points)
				// Minus the number of rune we have awarded them
				if vault.BondRewardRune.GTE(reward) {
					vault.BondRewardRune = vault.BondRewardRune.Sub(reward)
				} else {
					vault.BondRewardRune = sdk.ZeroUint()
				}
				k.SetVaultData(ctx, vault)
			}
		}
		na.ActiveBlockHeight = 0
	}
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(na))

	// When a node is in active status, we need to add the observer address to active
	// if it is not , then we could remove them
	if na.Status == NodeActive {
		k.SetActiveObserver(ctx, na.NodeAddress)
	} else {
		k.RemoveActiveObserver(ctx, na.NodeAddress)
	}
}

// Slash the bond of a node account
// NOTE: Should be careful not to slash too much, and have their Yggdrasil
// vault have more in funds than their bond. This could trigger them to have a
// untimely exit, stealing an amount of funds from stakers.
func (k KVStore) SlashNodeAccountBond(ctx sdk.Context, na *NodeAccount, slash sdk.Uint) {
	if slash.GT(na.Bond) {
		na.Bond = sdk.ZeroUint()
	} else {
		na.Bond = na.Bond.Sub(slash)
	}
	k.SetNodeAccount(ctx, *na)
}

// Slash the rewards of a node account
// NOTE: if we slash their rewards so much, they may do an orderly exit and
// rotate out of the active vault, wait in line to rejoin later.
func (k KVStore) SlashNodeAccountRewards(ctx sdk.Context, na *NodeAccount, pts int64) {
	na.SlashPoints += pts
	k.SetNodeAccount(ctx, *na)
}

func (k KVStore) EnsureTrustAccountUnique(ctx sdk.Context, consensusPubKey string, pubKeys common.PubKeys) error {
	iter := k.GetNodeAccountIterator(ctx)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var na NodeAccount
		if err := k.cdc.UnmarshalBinaryBare(iter.Value(), &na); nil != err {
			return errors.Wrap(err, "fail to unmarshal node account")
		}
		if na.ValidatorConsPubKey == consensusPubKey {
			return errors.Errorf("%s already exist", na.ValidatorConsPubKey)
		}
		if na.NodePubKey.Equals(pubKeys) {
			return errors.Errorf("%s already exist", pubKeys)
		}
	}

	return nil
}

// GetTrustAccountIterator iterate trust accounts
func (k KVStore) GetNodeAccountIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixNodeAccount))
}
