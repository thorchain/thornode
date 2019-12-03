package thorchain

import (
	"fmt"
	"strings"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperNodeAccount interface {
	TotalActiveNodeAccount(ctx sdk.Context) (int, error)
	ListNodeAccounts(ctx sdk.Context) (NodeAccounts, error)
	ListNodeAccountsByStatus(ctx sdk.Context, status NodeStatus) (NodeAccounts, error)
	ListActiveNodeAccounts(ctx sdk.Context) (NodeAccounts, error)
	GetLowestActiveVersion(ctx sdk.Context) semver.Version
	GetNodeAccount(ctx sdk.Context, addr sdk.AccAddress) (NodeAccount, error)
	GetNodeAccountByPubKey(ctx sdk.Context, pk common.PubKey) (NodeAccount, error)
	GetNodeAccountByBondAddress(ctx sdk.Context, addr common.Address) (NodeAccount, error)
	SetNodeAccount(ctx sdk.Context, na NodeAccount) error
	EnsureTrustAccountUnique(ctx sdk.Context, consensusPubKey string, pubKeys common.PubKeys) error
	GetNodeAccountIterator(ctx sdk.Context) sdk.Iterator
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
			return nodeAccounts, dbError(ctx, "Unmarshal: node account", err)
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
		return nodeAccounts, err
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
func (k KVStore) GetLowestActiveVersion(ctx sdk.Context) semver.Version {
	nodes, _ := k.ListActiveNodeAccounts(ctx)
	if len(nodes) > 0 {
		version := nodes[0].Version
		for _, na := range nodes {
			if na.Version.LT(version) {
				version = na.Version
			}
		}
		return version
	}
	return semver.Version{}
}

// GetNodeAccount try to get node account with the given address from db
func (k KVStore) GetNodeAccount(ctx sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	ctx.Logger().Debug("GetNodeAccount", "node account", addr.String())
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixNodeAccount, addr.String())
	payload := store.Get([]byte(key))
	var na NodeAccount
	if err := k.cdc.UnmarshalBinaryBare(payload, &na); nil != err {
		return na, dbError(ctx, "Unmarshal: node account", err)
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
		return na, err
	}
	for _, item := range nodeAccounts {
		if item.BondAddress.Equals(addr) {
			return item, nil
		}
	}
	return na, nil
}

// SetNodeAccount save the given node account into datastore
func (k KVStore) SetNodeAccount(ctx sdk.Context, na NodeAccount) error {
	ctx.Logger().Debug("SetNodeAccount", "node account", na.String())
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixNodeAccount, na.NodeAddress.String())
	if na.Status == NodeActive {
		if na.ActiveBlockHeight == 0 {
			// the na is active, and does not have a block height when they
			// became active. This must be the first block they are active, so
			// THORNode will set it now.
			na.ActiveBlockHeight = ctx.BlockHeight()
			na.SlashPoints = 0 // reset slash points
		}
	} else {
		if na.ActiveBlockHeight > 0 {

			// The node account seems to have become a non active node account.
			// Therefore, lets give them their bond rewards.
			vault, err := k.GetVaultData(ctx)
			if nil != err {
				return fmt.Errorf("fail to get vault: %w", err)
			}

			// Find number of blocks they have been an active node
			totalActiveBlocks := ctx.BlockHeight() - na.ActiveBlockHeight

			// find number of blocks they were well behaved (ie active - slash points)
			earnedBlocks := na.CalcBondUnits(ctx.BlockHeight())

			// calc number of rune they are awarded
			reward := vault.CalcNodeRewards(earnedBlocks)

			// Add to their bond the amount rewarded
			na.Bond = na.Bond.Add(reward)

			// Minus the number of rune THORNode have awarded them
			vault.BondRewardRune = common.SafeSub(vault.BondRewardRune, reward)

			// Minus the number of units na has (do not include slash points)
			vault.TotalBondUnits = common.SafeSub(
				vault.TotalBondUnits,
				sdk.NewUint(uint64(totalActiveBlocks)),
			)

			if err := k.SetVaultData(ctx, vault); nil != err {
				return fmt.Errorf("fail to save vault data: %w", err)
			}
		}
		na.ActiveBlockHeight = 0
	}
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(na))

	// When a node is in active status, THORNode need to add the observer address to active
	// if it is not , then THORNode could remove them
	if na.Status == NodeActive {
		k.SetActiveObserver(ctx, na.NodeAddress)
	} else {
		k.RemoveActiveObserver(ctx, na.NodeAddress)
	}
	return nil
}

func (k KVStore) EnsureTrustAccountUnique(ctx sdk.Context, consensusPubKey string, pubKeys common.PubKeys) error {
	iter := k.GetNodeAccountIterator(ctx)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var na NodeAccount
		if err := k.cdc.UnmarshalBinaryBare(iter.Value(), &na); nil != err {
			return dbError(ctx, "Unmarshal: node account", err)
		}
		if strings.EqualFold("", consensusPubKey) {
			return dbError(ctx, "", errors.New("Validator Consensus Key cannot be empty"))
		}
		if na.ValidatorConsPubKey == consensusPubKey {
			return dbError(ctx, "", errors.Errorf("%s already exist", na.ValidatorConsPubKey))
		}
		if pubKeys.Equals(common.EmptyPubKeys) {
			return dbError(ctx, "", errors.New("PubKeys cannot be empty"))
		}
		if na.NodePubKey.Equals(pubKeys) {
			return dbError(ctx, "", errors.Errorf("%s already exist", pubKeys))
		}
	}

	return nil
}

// GetTrustAccountIterator iterate trust accounts
func (k KVStore) GetNodeAccountIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixNodeAccount))
}
