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
	BeginBlock(ctx sdk.Context)
	EndBlock(ctx sdk.Context, store TxOutStore) []abci.ValidatorUpdate
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
func (vm *ValidatorMgr) BeginBlock(ctx sdk.Context) {
	height := ctx.BlockHeight()
	if height == genesisBlockHeight {
		if err := vm.setupValidatorNodes(ctx, height); nil != err {
			ctx.Logger().Error("fail to setup validator nodes", err)
		}
	}
}

// EndBlock when block end
func (vm *ValidatorMgr) EndBlock(ctx sdk.Context, store TxOutStore) []abci.ValidatorUpdate {
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
	for _, na := range activeNodes {
		found := false
		for _, member := range membership {
			if na.NodePubKey.Contains(member) {
				newActive = append(newActive, na)
				break
			}
		}
		if !found {
			na.UpdateStatus(NodeStandby, height)
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
				na.UpdateStatus(NodeActive, height)
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

	if len(newActive) <= constants.MinmumNodesForBFT { // THORNode still have enough validators for BFT
		// execute Ragnarok protocol, no going back
		// THORNode have to request the fund back now, because once it get to the rotate block height ,
		// THORNode won't have validators anymore
		if err := vm.ragnarokProtocolStep1(ctx, activeNodes, store); nil != err {
			ctx.Logger().Error("fail to execute ragnarok protocol step 1", err)
		}
	}

	return validators
}

func (vm *ValidatorMgr) RequestYggReturn(ctx sdk.Context, node NodeAccount, poolAddrMgr PoolAddressManager, txOut TxOutStore) error {
	ygg, err := vm.k.GetYggdrasil(ctx, node.NodePubKey.Secp256k1)
	if nil != err && !errors.Is(err, ErrYggdrasilNotFound) {
		return fmt.Errorf("fail to get yggdrasil: %w", err)
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
func (vm *ValidatorMgr) ragnarokProtocolStep1(ctx sdk.Context, activeNodes NodeAccounts, txOut TxOutStore) error {
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
func (vm *ValidatorMgr) setupValidatorNodes(ctx sdk.Context, height int64) error {
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
	for idx, item := range activeCandidateNodes {
		if int64(idx) < constants.DesireValidatorSet {
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
