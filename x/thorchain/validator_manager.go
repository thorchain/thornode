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

// ValidatorManager is to manage a list of validators , and rotate them
type ValidatorManager struct {
	k              Keeper
	Meta           *ValidatorMeta
	rotationPolicy ValidatorRotationPolicy
	poolAddrMgr    *PoolAddressManager
}

// NewValidatorManager create a new instance of ValidatorManager
func NewValidatorManager(k Keeper, poolAddrMgr *PoolAddressManager) *ValidatorManager {
	return &ValidatorManager{
		k:           k,
		poolAddrMgr: poolAddrMgr,
	}
}

// BeginBlock when block begin
func (vm *ValidatorManager) BeginBlock(ctx sdk.Context) {
	height := ctx.BlockHeight()
	vm.rotationPolicy = GetValidatorRotationPolicy(ctx, vm.k)
	if err := vm.rotationPolicy.IsValid(); nil != err {
		ctx.Logger().Error("invalid rotation policy", err)
	}
	if height == genesisBlockHeight {
		vm.Meta = &ValidatorMeta{}
		if err := vm.setupValidatorNodes(ctx, height); nil != err {
			ctx.Logger().Error("fail to setup validator nodes", err)
		}
		vm.k.SetValidatorMeta(ctx, *vm.Meta)
	}

	// restore vm.meta from data store
	if vm.Meta == nil {
		meta := vm.k.GetValidatorMeta(ctx)
		vm.Meta = &meta
	}
}

// EndBlock when block end
func (vm *ValidatorManager) EndBlock(ctx sdk.Context, store *TxOutStore) []abci.ValidatorUpdate {
	height := ctx.BlockHeight()
	if height != vm.Meta.RotateWindowOpenAtBlockHeight &&
		height != vm.Meta.RotateAtBlockHeight &&
		height != vm.Meta.LeaveOpenWindow &&
		height != vm.Meta.LeaveProcessAt {
		return nil
	}
	defer func() {
		vm.k.SetValidatorMeta(ctx, *vm.Meta)
	}()
	if height == vm.Meta.RotateWindowOpenAtBlockHeight {
		if err := vm.prepareAddNode(ctx, height); nil != err {
			ctx.Logger().Error("fail to prepare add nodes", err)
		}
		return nil
	}

	if height == vm.Meta.LeaveOpenWindow {
		if err := vm.prepareToNodesToLeave(ctx, store); nil != err {
			ctx.Logger().Error("fail to prepare node to leave")
		}
		return nil
	}
	if height == vm.Meta.RotateAtBlockHeight || height == vm.Meta.LeaveProcessAt {
		//  queueNode put in here on purpose, so THORNode know who had been queued before THORNode actually rotate them
		queueNode := vm.Meta.Queued
		rotated := false
		var err error
		if height == vm.Meta.RotateAtBlockHeight {
			rotated, err = vm.rotateValidatorNodes(ctx, store)
			if nil != err {
				ctx.Logger().Error("fail to rotate validator nodes", err)
				return nil
			}
		}
		if height == vm.Meta.LeaveProcessAt {
			rotated, err = vm.processValidatorLeave(ctx, store)
			if nil != err {
				ctx.Logger().Error("fail to process validator leave", err)
				return nil
			}
		}
		if !rotated {
			return nil
		}

		activeNodes, err := vm.k.ListActiveNodeAccounts(ctx)
		if nil != err {
			ctx.Logger().Error("fail to list all active node accounts")
			return nil
		}
		validators := make([]abci.ValidatorUpdate, 0, len(activeNodes))
		for _, item := range activeNodes {
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
		for _, item := range queueNode {
			na, err := vm.k.GetNodeAccount(ctx, item.NodeAddress)
			if nil != err {
				ctx.Logger().Error("fail to get node account", err, "node address", item.NodeAddress.String())
				continue
			}
			if na.Status == NodeStandby {
				// node to be removed as validator
				pk, err := sdk.GetConsPubKeyBech32(na.ValidatorConsPubKey)
				if nil != err {
					ctx.Logger().Error("fail to parse consensus public key", "key", na.ValidatorConsPubKey)
					continue
				}
				validators = append(validators, abci.ValidatorUpdate{
					PubKey: tmtypes.TM2PB.PubKey(pk),
					Power:  0,
				})
			}
		}
		return validators
	}

	return nil
}
func (vm *ValidatorManager) processValidatorLeave(ctx sdk.Context, store *TxOutStore) (bool, error) {
	if vm.Meta.LeaveProcessAt != ctx.BlockHeight() {
		return false, nil
	}
	defer func() {
		vm.Meta.Nominated = nil
		vm.Meta.Queued = nil
		vm.Meta.LeaveQueue = nil
		vm.Meta.LeaveOpenWindow += vm.rotationPolicy.LeaveProcessPerBlockHeight
		vm.Meta.LeaveProcessAt += vm.rotationPolicy.LeaveProcessPerBlockHeight
		// delay scheduled validator rotation
		vm.Meta.RotateAtBlockHeight += vm.rotationPolicy.LeaveProcessPerBlockHeight
		vm.Meta.RotateWindowOpenAtBlockHeight += vm.rotationPolicy.LeaveProcessPerBlockHeight
	}()

	// Ragnarok protocol
	if vm.Meta.Ragnarok {
		ctx.Logger().Info("Ragnarok protocol triggered")
		return false, nil
	}

	if vm.Meta.Queued.IsEmpty() {
		ctx.Logger().Info("no nodes need to leave so no rotate")
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeValidatorManager,
				sdk.NewAttribute("action", "abort"),
				sdk.NewAttribute("reason", "no queued nodes")))
		return false, nil
	}

	for _, item := range vm.Meta.Queued {
		item.UpdateStatus(NodeStandby, ctx.BlockHeight())
		vm.k.SetNodeAccount(ctx, item)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeValidatorStandby,
				sdk.NewAttribute("bep_address", item.NodeAddress.String()),
				sdk.NewAttribute("consensus_public_key", item.ValidatorConsPubKey)))
		if err := vm.requestYggReturn(ctx, item, vm.poolAddrMgr, store); nil != err {
			return false, err
		}
	}

	for _, item := range vm.Meta.Nominated {
		nominatedNodeAccount, err := vm.k.GetNodeAccount(ctx, item.NodeAddress)
		if nil != err {
			return false, fmt.Errorf("fail to get nominated account from data store: %w", err)
		}

		if nominatedNodeAccount.Status != NodeReady {
			// set them to standby, do THORNode need to slash the validator? THORNode nominated them but they are not ready
			nominatedNodeAccount.UpdateStatus(NodeStandby, ctx.BlockHeight())
			vm.k.SetNodeAccount(ctx, nominatedNodeAccount)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(EventTypeValidatorManager,
					sdk.NewAttribute("bep_address", nominatedNodeAccount.NodeAddress.String()),
					sdk.NewAttribute("consensus_public_key", nominatedNodeAccount.ValidatorConsPubKey),
					sdk.NewAttribute("action", "abort"),
					sdk.NewAttribute("reason", "node not ready")))
			ctx.Logger().Info(fmt.Sprintf("nominated account %s is not ready , abort rotation, try it nex time", item.NodeAddress))
			// go to the next
			continue
		}
		// set to active
		nominatedNodeAccount.UpdateStatus(NodeActive, ctx.BlockHeight())
		vm.k.SetNodeAccount(ctx, nominatedNodeAccount)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeValidatorActive,
				sdk.NewAttribute("bep_address", nominatedNodeAccount.NodeAddress.String()),
				sdk.NewAttribute("consensus_public_key", nominatedNodeAccount.ValidatorConsPubKey)))
	}
	return true, nil
}
func (vm *ValidatorManager) rotateValidatorNodes(ctx sdk.Context, store *TxOutStore) (bool, error) {
	if vm.Meta.RotateAtBlockHeight != ctx.BlockHeight() {
		// it is not an error , just not a good time
		return false, nil
	}
	defer func() {
		vm.Meta.Nominated = nil
		vm.Meta.Queued = nil
		vm.Meta.LeaveQueue = nil
		vm.Meta.LeaveOpenWindow += vm.rotationPolicy.LeaveProcessPerBlockHeight
		vm.Meta.LeaveProcessAt += vm.rotationPolicy.LeaveProcessPerBlockHeight
		vm.Meta.RotateAtBlockHeight += vm.rotationPolicy.RotatePerBlockHeight
		vm.Meta.RotateWindowOpenAtBlockHeight += vm.rotationPolicy.RotatePerBlockHeight
	}()
	// Ragnarok protocol
	if vm.Meta.Ragnarok {
		ctx.Logger().Info("Ragnarok protocol triggered , no more rotation")
		return false, nil
	}

	if vm.Meta.Nominated.IsEmpty() {
		ctx.Logger().Info("no nodes get nominated ,and no nodes need to leave so no rotate")
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeValidatorManager,
				sdk.NewAttribute("action", "abort"),
				sdk.NewAttribute("reason", "no nominated nodes")))
		return false, nil
	}
	hasRotateIn := false
	for _, item := range vm.Meta.Nominated {
		nominatedNodeAccount, err := vm.k.GetNodeAccount(ctx, item.NodeAddress)
		if nil != err {
			return false, fmt.Errorf("fail to get nominated account from data store: %w", err)
		}

		if nominatedNodeAccount.Status != NodeReady {
			// set them to standby, do THORNode need to slash the validator? THORNode nominated them but they are not ready
			nominatedNodeAccount.UpdateStatus(NodeStandby, ctx.BlockHeight())
			vm.k.SetNodeAccount(ctx, nominatedNodeAccount)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(EventTypeValidatorManager,
					sdk.NewAttribute("bep_address", nominatedNodeAccount.NodeAddress.String()),
					sdk.NewAttribute("consensus_public_key", nominatedNodeAccount.ValidatorConsPubKey),
					sdk.NewAttribute("action", "abort"),
					sdk.NewAttribute("reason", "node not ready")))
			ctx.Logger().Info(fmt.Sprintf("nominated account %s is not ready , abort rotation, try it nex time", item.NodeAddress))
			// go to the next
			continue
		}
		// set to active
		nominatedNodeAccount.UpdateStatus(NodeActive, ctx.BlockHeight())
		vm.k.SetNodeAccount(ctx, nominatedNodeAccount)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeValidatorActive,
				sdk.NewAttribute("bep_address", nominatedNodeAccount.NodeAddress.String()),
				sdk.NewAttribute("consensus_public_key", nominatedNodeAccount.ValidatorConsPubKey)))
		hasRotateIn = true
	}
	if !hasRotateIn {
		return false, errors.New("none of the nominated node is ready, give up")
	}

	if !vm.Meta.Queued.IsEmpty() {
		for _, item := range vm.Meta.Queued {
			item.UpdateStatus(NodeStandby, ctx.BlockHeight())
			vm.k.SetNodeAccount(ctx, item)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(EventTypeValidatorStandby,
					sdk.NewAttribute("bep_address", item.NodeAddress.String()),
					sdk.NewAttribute("consensus_public_key", item.ValidatorConsPubKey)))
			// request money back
			if err := vm.requestYggReturn(ctx, item, vm.poolAddrMgr, store); nil != err {
				return false, err
			}
		}
	}
	return true, nil
}

func (vm *ValidatorManager) requestYggReturn(ctx sdk.Context, node NodeAccount, poolAddrMgr *PoolAddressManager, txOut *TxOutStore) error {
	ygg := vm.k.GetYggdrasil(ctx, node.NodePubKey.Secp256k1)
	chains, err := vm.k.GetChains(ctx)
	if err != nil {
		return err
	}
	if !ygg.HasFunds() {
		return nil
	}
	for _, c := range chains {
		currentChainPoolAddr := poolAddrMgr.currentPoolAddresses.Current.GetByChain(c)
		for _, coin := range ygg.Coins {
			toAddr, err := currentChainPoolAddr.PubKey.GetAddress(coin.Asset.Chain)
			if !toAddr.IsEmpty() {
				txOutItem := &TxOutItem{
					Chain:       coin.Asset.Chain,
					ToAddress:   toAddr,
					InHash:      common.BlankTxID,
					PoolAddress: ygg.PubKey,
					Memo:        "yggdrasil-",
					Coin:        coin,
				}
				txOut.AddTxOutItem(ctx, vm.k, txOutItem, false)
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

// leave process window open
func (vm *ValidatorManager) prepareToNodesToLeave(ctx sdk.Context, txOut *TxOutStore) error {
	height := ctx.BlockHeight()
	if height != vm.Meta.LeaveOpenWindow {
		return nil
	}
	if len(vm.Meta.LeaveQueue) == 0 {
		ctx.Logger().Info("no one request to leave")
		return nil
	}

	// honour leave request
	for _, item := range vm.Meta.LeaveQueue {
		node, err := vm.k.GetNodeAccount(ctx, item.NodeAddress)
		if nil != err {
			return fmt.Errorf("fail to get node account: %w", err)
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeQueuedValidator,
				sdk.NewAttribute("bep_address", node.NodeAddress.String()),
				sdk.NewAttribute("consensus_public_key", node.ValidatorConsPubKey)))
		vm.Meta.Queued = append(vm.Meta.Queued, node)
	}
	rotateIn := len(vm.Meta.Queued)
	// do THORNode have standby nodes?
	// who should be added , and who need to removed
	standbyNodes, err := vm.k.ListNodeAccountsByStatus(ctx, NodeStandby)
	if nil != err {
		return fmt.Errorf("fail to get all standby nodes,%w", err)
	}
	if len(standbyNodes) == 0 {
		ctx.Logger().Info("no standby nodes")
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeValidatorManager,
				sdk.NewAttribute("action", "ignore"),
				sdk.NewAttribute("reason", "no standby nodes")))
	}

	if len(standbyNodes) > rotateIn {
		sort.Sort(standbyNodes)
		standbyNodes = standbyNodes[:rotateIn]
	}

	vm.Meta.Nominated = standbyNodes
	for _, item := range vm.Meta.Nominated {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeNominatedValidator,
				sdk.NewAttribute("bep_address", item.NodeAddress.String()),
				sdk.NewAttribute("consensus_public_key", item.ValidatorConsPubKey)))
	}

	activeNodes, err := vm.k.ListActiveNodeAccounts(ctx)
	if nil != err {
		return fmt.Errorf("fail to get active node accounts: %w", err)
	}
	totalActive := len(activeNodes)
	afterLeave := totalActive + len(vm.Meta.Nominated) - len(vm.Meta.Queued)

	if afterLeave <= constants.MinmumNodesForYggdrasil {
		if err := vm.recallYggFunds(ctx, activeNodes, txOut); nil != err {
			return fmt.Errorf("fail to recall yggdrasil funds")
		}
	}

	if afterLeave > constants.MinmumNodesForBFT { // THORNode still have enough validators for BFT
		// trigger pool rotate next
		vm.poolAddrMgr.currentPoolAddresses.RotateWindowOpenAt = height + 1
		vm.poolAddrMgr.currentPoolAddresses.RotateAt = vm.Meta.LeaveProcessAt
		return nil
	}
	// execute Ragnarok protocol, no going back
	// THORNode have to request the fund back now, because once it get to the rotate block height ,
	// THORNode won't have validators anymore
	if err := vm.ragnarokProtocolStep1(ctx, activeNodes, txOut); nil != err {
		return fmt.Errorf("fail to execute ragnarok protocol step 1")
	}

	return nil
}

// ragnarokProtocolStep1 - request all yggdrasil pool to return the fund
// when THORNode observe the node return fund successfully, the node's bound will be refund.
func (vm *ValidatorManager) ragnarokProtocolStep1(ctx sdk.Context, activeNodes NodeAccounts, txOut *TxOutStore) error {
	vm.Meta.Ragnarok = true
	// do THORNode have yggdrasil pool?
	hasYggdrasil, err := vm.k.HasValidYggdrasilPools(ctx)
	if nil != err {
		return fmt.Errorf("fail at ragnarok protocol step 1: %w", err)
	}
	if !hasYggdrasil {
		result := handleRagnarokProtocolStep2(ctx, vm.k, txOut, vm.poolAddrMgr, vm)
		if !result.IsOK() {
			return errors.New("fail to process ragnarok protocol step 2")
		}
		return nil
	}
	return vm.recallYggFunds(ctx, activeNodes, txOut)
}

func (vm *ValidatorManager) recallYggFunds(ctx sdk.Context, activeNodes NodeAccounts, txOut *TxOutStore) error {
	// request every node to return fund
	for _, na := range activeNodes {
		if err := vm.requestYggReturn(ctx, na, vm.poolAddrMgr, txOut); nil != err {
			return fmt.Errorf("fail to request yggdrasil fund back: %w", err)
		}
	}
	return nil
}

// scheduled node rotation open
func (vm *ValidatorManager) prepareAddNode(ctx sdk.Context, height int64) error {
	if height != vm.Meta.RotateWindowOpenAtBlockHeight {
		// it is not an error , just not a good time to do this yet
		return nil
	}
	// who should be added , and who need to removed
	standbyNodes, err := vm.k.ListNodeAccountsByStatus(ctx, NodeStandby)
	if nil != err {
		return fmt.Errorf("fail to get all standby nodes,%w", err)
	}
	if len(standbyNodes) == 0 {
		ctx.Logger().Info("no standby nodes")
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeValidatorManager,
				sdk.NewAttribute("action", "abort"),
				sdk.NewAttribute("reason", "no standby nodes")))
		return nil
	}
	sort.Sort(standbyNodes)
	activeNodes, err := vm.k.ListActiveNodeAccounts(ctx)
	if nil != err {
		return fmt.Errorf("fail to get active node accounts: %w", err)
	}
	totalActiveNodes := len(activeNodes)
	rotateIn := vm.rotationPolicy.RotateInNumBeforeFull
	rotateOut := vm.rotationPolicy.RotateOutNumBeforeFull
	if int64(totalActiveNodes) >= vm.rotationPolicy.DesireValidatorSet {
		// THORNode are full
		rotateIn = vm.rotationPolicy.RotateNumAfterFull
		rotateOut = vm.rotationPolicy.RotateNumAfterFull
	}
	if int64(len(standbyNodes)) > rotateIn {
		standbyNodes = standbyNodes[:rotateIn]
	}
	vm.Meta.Nominated = standbyNodes
	for _, item := range vm.Meta.Nominated {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(EventTypeNominatedValidator,
				sdk.NewAttribute("bep_address", item.NodeAddress.String()),
				sdk.NewAttribute("consensus_public_key", item.ValidatorConsPubKey)))
	}
	// THORNode need to set a minimum validator set , if we have less than the minimum , then THORNode don't rotate out
	if totalActiveNodes <= constants.MinmumNodesForBFT {
		rotateOut = 0
	}

	if rotateOut > 0 {
		mod := (ctx.BlockHeight() / vm.rotationPolicy.RotatePerBlockHeight) % 2
		if mod == 0 {
			activeNodesBySlash := NodeAccountsBySlashingPoint(activeNodes)
			sort.Sort(activeNodesBySlash)
			// Queue the first few nodes to be rotated out
			vm.Meta.Queued = NodeAccounts(activeNodesBySlash[:rotateOut])
		} else {
			sort.Sort(activeNodes)
			vm.Meta.Queued = activeNodes[:rotateOut]
		}
		for _, item := range vm.Meta.Queued {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(EventTypeQueuedValidator,
					sdk.NewAttribute("bep_address", item.NodeAddress.String()),
					sdk.NewAttribute("consensus_public_key", item.ValidatorConsPubKey)))
		}
	}
	return nil
}

// setupValidatorNodes it is one off it only get called when genesis
func (vm *ValidatorManager) setupValidatorNodes(ctx sdk.Context, height int64) error {
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
			ctx.Logger().Error("fail to unmarshal node account", "error", err)
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
	desireValidatorSet := vm.rotationPolicy.DesireValidatorSet
	for idx, item := range activeCandidateNodes {
		if int64(idx) < desireValidatorSet {
			item.UpdateStatus(NodeActive, ctx.BlockHeight())
		} else {
			item.UpdateStatus(NodeStandby, ctx.BlockHeight())
		}
		vm.k.SetNodeAccount(ctx, item)
	}
	vm.Meta.RotateAtBlockHeight = vm.rotationPolicy.RotatePerBlockHeight + 1
	vm.Meta.RotateWindowOpenAtBlockHeight = vm.rotationPolicy.RotatePerBlockHeight + 1 - vm.rotationPolicy.ValidatorChangeWindow
	vm.Meta.LeaveOpenWindow = vm.rotationPolicy.LeaveProcessPerBlockHeight + 1 - vm.rotationPolicy.ValidatorChangeWindow
	vm.Meta.LeaveProcessAt = vm.rotationPolicy.LeaveProcessPerBlockHeight + 1
	return nil
}
