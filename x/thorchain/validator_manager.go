package thorchain

import (
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

const (
	EventTypeValidatorManager   = `validator_manager`
	EventTypeNominatedValidator = `validator_nominated`
	EventTypeQueuedValidator    = `validator_queued`
	EventTypeValidatorActive    = `validator_active`
	EventTypeValidatorStandby   = `validator_standby`

	genesisBlockHeight = 1
)

type ValidatorManager interface {
	BeginBlock(ctx sdk.Context, constAccessor constants.ConstantValues) error
	EndBlock(ctx sdk.Context, store TxOutStore, constAccessor constants.ConstantValues) []abci.ValidatorUpdate
	RequestYggReturn(ctx sdk.Context, node NodeAccount, poolAddrMgr PoolAddressManager, txOut TxOutStore) error
}

// ValidatorMgr is to manage a list of validators , and rotate them
type ValidatorMgr struct {
	k           Keeper
	poolAddrMgr PoolAddressManager
}

// NewValidatorManager create a new instance of ValidatorManager
func NewValidatorMgr(k Keeper, poolAddrMgr PoolAddressManager) *ValidatorMgr {
	return &ValidatorMgr{
		k:           k,
		poolAddrMgr: poolAddrMgr,
	}
}

// BeginBlock when block begin
func (vm *ValidatorMgr) BeginBlock(ctx sdk.Context, constAccessor constants.ConstantValues) error {
	height := ctx.BlockHeight()
	if height == genesisBlockHeight {
		if err := vm.setupValidatorNodes(ctx, height, constAccessor); nil != err {
			ctx.Logger().Error("fail to setup validator nodes", err)
		}
	}
	badValidatorRate := constAccessor.GetInt64Value(constants.BadValidatorRate)
	if err := vm.markBadActor(ctx, badValidatorRate); err != nil {
		return err
	}
	oldValidatorRate := constAccessor.GetInt64Value(constants.OldValidatorRate)
	if err := vm.markOldActor(ctx, oldValidatorRate); err != nil {
		return err
	}

	rotatePerBlockHeight := constAccessor.GetInt64Value(constants.RotatePerBlockHeight)
	desireValidatorSet := constAccessor.GetInt64Value(constants.DesireValidatorSet)
	if ctx.BlockHeight()%rotatePerBlockHeight == 0 {
		next, ok, err := vm.nextPoolNodeAccounts(ctx, int(desireValidatorSet))
		if err != nil {
			return err
		}
		if ok {
			keygen := make(Keygen, len(next))
			for i := range next {
				keygen[i] = next[i].NodePubKey.Secp256k1
			}
			keygens := NewKeygens(uint64(ctx.BlockHeight()))
			keygens.Keygens = []Keygen{keygen}
			if err := vm.k.SetKeygens(ctx, keygens); err != nil {
				return err
			}
		}
	}

	return nil
}

// EndBlock when block end
func (vm *ValidatorMgr) EndBlock(ctx sdk.Context, store TxOutStore, constAccessor constants.ConstantValues) []abci.ValidatorUpdate {
	height := ctx.BlockHeight()
	activeNodes, err := vm.k.ListActiveNodeAccounts(ctx)
	if err != nil {
		ctx.Logger().Error("fail to list active node accounts")
	}

	readyNodes, err := vm.k.ListNodeAccountsByStatus(ctx, NodeReady)
	if err != nil {
		ctx.Logger().Error("fail to list ready node accounts")
	}

	poolAddresses, err := vm.k.GetPoolAddresses(ctx)
	if err != nil {
		ctx.Logger().Error("fail to get pool addresses")
	}
	membership := poolAddresses.Current[0].Membership

	var newActive NodeAccounts // store the list of new active users

	// find active node accounts that are no longer active
	removedNodes := false
	for _, na := range activeNodes {
		found := false
		for _, member := range membership {
			if na.NodePubKey.Contains(member) {
				newActive = append(newActive, na)
				na.TryAddSignerPubKey(poolAddresses.Current[0].PubKey)
				if err := vm.k.SetNodeAccount(ctx, na); err != nil {
					ctx.Logger().Error("fail to save node account")
				}
				break
			}
		}
		if !found && len(membership) > 0 {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent("UpdateNodeAccountStatus",
					sdk.NewAttribute("Address", na.NodeAddress.String()),
					sdk.NewAttribute("Former:", na.Status.String()),
					sdk.NewAttribute("Current:", NodeStandby.String())))
			na.UpdateStatus(NodeStandby, height)
			removedNodes = true
			if err := vm.k.SetNodeAccount(ctx, na); err != nil {
				ctx.Logger().Error("fail to save node account")
			}
		}
	}

	// find ready nodes that change to
	for _, na := range readyNodes {
		for _, member := range membership {
			if na.NodePubKey.Contains(member) {
				newActive = append(newActive, na)
				ctx.EventManager().EmitEvent(
					sdk.NewEvent("UpdateNodeAccountStatus",
						sdk.NewAttribute("Address", na.NodeAddress.String()),
						sdk.NewAttribute("Former:", na.Status.String()),
						sdk.NewAttribute("Current:", NodeActive.String())))
				na.UpdateStatus(NodeActive, height)
				na.TryAddSignerPubKey(poolAddresses.Current[0].PubKey)
				if err := vm.k.SetNodeAccount(ctx, na); err != nil {
					ctx.Logger().Error("fail to save node account")
				}
				break
			}
		}
	}

	validators := make([]abci.ValidatorUpdate, 0, len(newActive))
	for _, item := range newActive {
		pk, err := sdk.GetConsPubKeyBech32(item.ValidatorConsPubKey)
		if nil != err {
			ctx.Logger().Error("fail to parse consensus public key", "key", item.ValidatorConsPubKey)
			continue
		}
		validators = append(validators, abci.ValidatorUpdate{
			PubKey: tmtypes.TM2PB.PubKey(pk),
			Power:  100,
		})
	}
	minimumNodesForBFT := constAccessor.GetInt64Value(constants.MinimumNodesForBFT)
	if height > 1 && removedNodes && len(newActive) <= int(minimumNodesForBFT) { // THORNode still have enough validators for BFT
		// execute Ragnarok protocol, no going back
		// THORNode have to request the fund back now, because once it get to the rotate block height ,
		// THORNode won't have validators anymore
		if err := vm.ragnarokProtocolStep1(ctx, activeNodes, store, constAccessor); nil != err {
			ctx.Logger().Error("fail to execute ragnarok protocol step 1: %s", err)
		}
	}

	return validators
}

func (vm *ValidatorMgr) RequestYggReturn(ctx sdk.Context, node NodeAccount, poolAddrMgr PoolAddressManager, txOut TxOutStore) error {
	ygg, err := vm.k.GetVault(ctx, node.NodePubKey.Secp256k1)
	if nil != err {
		return fmt.Errorf("fail to get yggdrasil: %w", err)
	}
	if !ygg.IsYggdrasil() {
		return fmt.Errorf("this is not a Yggdrasil vault")
	}
	chains, err := vm.k.GetChains(ctx)
	if err != nil {
		return err
	}
	if !ygg.HasFunds() {
		return nil
	}
	for _, c := range chains {
		currentChainPoolAddr := poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(c)
		for _, coin := range ygg.Coins {
			toAddr, err := currentChainPoolAddr.PubKey.GetAddress(coin.Asset.Chain)
			if !toAddr.IsEmpty() {
				txOutItem := &TxOutItem{
					Chain:       coin.Asset.Chain,
					ToAddress:   toAddr,
					InHash:      common.BlankTxID,
					VaultPubKey: ygg.PubKey,
					Memo:        "yggdrasil-",
					Coin:        coin,
				}
				txOut.AddTxOutItem(ctx, txOutItem)
				continue
			}
			wrapErr := fmt.Errorf(
				"fail to get pool address (%s) for chain (%s) : %w",
				toAddr.String(),
				coin.Asset.Chain.String(), err)
			ctx.Logger().Error(wrapErr.Error(), "error", err)
			return wrapErr
		}
	}
	return nil
}

// ragnarokProtocolStep1 - request all yggdrasil pool to return the fund
// when THORNode observe the node return fund successfully, the node's bound will be refund.
func (vm *ValidatorMgr) ragnarokProtocolStep1(ctx sdk.Context, activeNodes NodeAccounts, txOut TxOutStore, constAccessor constants.ConstantValues) error {
	// do THORNode have yggdrasil pool?
	hasYggdrasil, err := vm.k.HasValidVaultPools(ctx)
	if nil != err {
		return fmt.Errorf("fail at ragnarok protocol step 1: %w", err)
	}
	if !hasYggdrasil {
		result := handleRagnarokProtocolStep2(ctx, vm.k, txOut, vm.poolAddrMgr, constAccessor)
		if !result.IsOK() {
			return errors.New("fail to process ragnarok protocol step 2")
		}
		return nil
	}
	return vm.recallYggFunds(ctx, activeNodes, txOut)
}

func (vm *ValidatorMgr) recallYggFunds(ctx sdk.Context, activeNodes NodeAccounts, txOut TxOutStore) error {
	// request every node to return fund
	for _, na := range activeNodes {
		if err := vm.RequestYggReturn(ctx, na, vm.poolAddrMgr, txOut); nil != err {
			return fmt.Errorf("fail to request yggdrasil fund back: %w", err)
		}
	}
	return nil
}

// setupValidatorNodes it is one off it only get called when genesis
func (vm *ValidatorMgr) setupValidatorNodes(ctx sdk.Context, height int64, constAccessor constants.ConstantValues) error {
	if height != genesisBlockHeight {
		ctx.Logger().Info("only need to setup validator node when start up", "height", height)
		return nil
	}

	iter := vm.k.GetNodeAccountIterator(ctx)
	defer iter.Close()
	readyNodes := NodeAccounts{}
	activeCandidateNodes := NodeAccounts{}
	for ; iter.Valid(); iter.Next() {
		var na NodeAccount
		if err := vm.k.Cdc().UnmarshalBinaryBare(iter.Value(), &na); nil != err {
			return fmt.Errorf("fail to unmarshal node account, %w", err)
		}
		// when THORNode first start , THORNode only care about these two status
		switch na.Status {
		case NodeReady:
			readyNodes = append(readyNodes, na)
		case NodeActive:
			activeCandidateNodes = append(activeCandidateNodes, na)
		}
	}
	totalActiveValidators := len(activeCandidateNodes)
	totalNominatedValidators := len(readyNodes)
	if totalActiveValidators == 0 && totalNominatedValidators == 0 {
		return errors.New("no validators available")
	}

	sort.Sort(activeCandidateNodes)
	sort.Sort(readyNodes)
	activeCandidateNodes = append(activeCandidateNodes, readyNodes...)
	desireValidatorSet := constAccessor.GetInt64Value(constants.DesireValidatorSet)
	for idx, item := range activeCandidateNodes {
		if int64(idx) < desireValidatorSet {
			item.UpdateStatus(NodeActive, ctx.BlockHeight())
		} else {
			item.UpdateStatus(NodeStandby, ctx.BlockHeight())
		}
		if err := vm.k.SetNodeAccount(ctx, item); nil != err {
			return fmt.Errorf("fail to save node account: %w", err)
		}
	}
	return nil
}

// Iterate over active node accounts, finding the one with the most slash points
func (vm *ValidatorMgr) findBadActor(ctx sdk.Context) (NodeAccount, error) {
	na := NodeAccount{}
	nas, err := vm.k.ListActiveNodeAccounts(ctx)
	if err != nil {
		return na, err
	}

	// TODO: return if we're at risk of loosing BTF

	// Find bad actor relative to slashpoints / age.
	// NOTE: avoiding the usage of float64, we use an alt method...
	na.SlashPoints = 1
	na.StatusSince = 9223372036854775807 // highest int64 value
	for _, n := range nas {
		if n.SlashPoints == 0 {
			continue
		}

		naVal := n.StatusSince / na.SlashPoints
		nVal := n.StatusSince / n.SlashPoints
		if nVal > (naVal) {
			na = n
		} else if nVal == naVal {
			if n.SlashPoints > na.SlashPoints {
				na = n
			}
		}
	}

	return na, nil
}

// Iterate over active node accounts, finding the one that has been active longest
func (vm *ValidatorMgr) findOldActor(ctx sdk.Context) (NodeAccount, error) {
	na := NodeAccount{}
	nas, err := vm.k.ListActiveNodeAccounts(ctx)
	if err != nil {
		return na, err
	}

	// TODO: return if we're at risk of loosing BTF

	na.StatusSince = ctx.BlockHeight() // set the start status age to "now"
	for _, n := range nas {
		if n.StatusSince < na.StatusSince {
			na = n
		}
	}

	return na, nil
}

// Mark an old to be churned out
func (vm *ValidatorMgr) markActor(ctx sdk.Context, na NodeAccount) error {
	if na.LeaveHeight == 0 {
		na.LeaveHeight = ctx.BlockHeight()
		return vm.k.SetNodeAccount(ctx, na)
	}
	return nil
}

// Mark an old actor to be churned out
func (vm *ValidatorMgr) markOldActor(ctx sdk.Context, rate int64) error {
	if rate%ctx.BlockHeight() == 0 {
		na, err := vm.findOldActor(ctx)
		if err != nil {
			return err
		}
		if err := vm.markActor(ctx, na); err != nil {
			return err
		}
	}
	return nil
}

// Mark a bad actor to be churned out
func (vm *ValidatorMgr) markBadActor(ctx sdk.Context, rate int64) error {
	if rate%ctx.BlockHeight() == 0 {
		na, err := vm.findBadActor(ctx)
		if err != nil {
			return err
		}
		if err := vm.markActor(ctx, na); err != nil {
			return err
		}
	}
	return nil
}

// find any actor that are ready to become "ready" status
func (vm *ValidatorMgr) markReadyActors(ctx sdk.Context) error {
	standby, err := vm.k.ListNodeAccountsByStatus(ctx, NodeStandby)
	if err != nil {
		return err
	}
	ready, err := vm.k.ListNodeAccountsByStatus(ctx, NodeReady)
	if err != nil {
		return err
	}

	// find min version node has to be, to be "ready" status
	minVersion := vm.k.GetMinJoinVersion(ctx)

	// check all ready and standby nodes are in "ready" state (upgrade/downgrade as needed)
	for _, na := range append(standby, ready...) {
		na.Status = NodeReady // everyone starts with the benefit of the doubt

		// TODO: check node is up to date on thorchain, binance, etc
		// must have made an observation that matched 2/3rds within the last 5 blocks

		// Check version number is still supported
		if na.Version.LT(minVersion) {
			na.UpdateStatus(NodeStandby, ctx.BlockHeight())
		}

		if err := vm.k.SetNodeAccount(ctx, na); err != nil {
			return err
		}
	}

	return nil
}

// Returns a list of nodes to include in the next pool
func (vm *ValidatorMgr) nextPoolNodeAccounts(ctx sdk.Context, targetCount int) (NodeAccounts, bool, error) {
	rotation := false // track if are making any changes to the current active node accounts

	// update list of ready actors
	if err := vm.markReadyActors(ctx); err != nil {
		return nil, false, err
	}

	ready, err := vm.k.ListNodeAccountsByStatus(ctx, NodeReady)
	if err != nil {
		return nil, false, err
	}
	// sort by bond size
	sort.Slice(ready, func(i, j int) bool {
		return ready[i].Bond.GT(ready[j].Bond)
	})

	active, err := vm.k.ListActiveNodeAccounts(ctx)
	if err != nil {
		return nil, false, err
	}
	// sort by LeaveHeight, giving preferential treatment to people who
	// requested to leave
	sort.Slice(active, func(i, j int) bool {
		if active[i].RequestedToLeave != active[j].RequestedToLeave {
			return active[i].RequestedToLeave
		}
		return active[i].LeaveHeight < active[j].LeaveHeight
	})

	// remove a node node account, if one is marked to leave
	if len(active) > 0 && active[0].LeaveHeight > 0 {
		rotation = true
		active = active[1:]
	}

	// add ready nodes to become active
	limit := 2 // Max limit of ready nodes to add. TODO: this should be a constant
	for i := 1; i <= targetCount-len(active); i++ {
		if len(ready) >= i {
			rotation = true
			active = append(active, ready[i-1])
		}
		if i == limit { // limit adding ready accounts
			break
		}
	}

	return active, rotation, nil
}
