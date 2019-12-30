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
	genesisBlockHeight = 1
)

type ValidatorManager interface {
	BeginBlock(ctx sdk.Context, constAccessor constants.ConstantValues) error
	EndBlock(ctx sdk.Context, constAccessor constants.ConstantValues) []abci.ValidatorUpdate
	RequestYggReturn(ctx sdk.Context, node NodeAccount) error
}

// ValidatorMgr is to manage a list of validators , and rotate them
type ValidatorMgr struct {
	k          Keeper
	vaultMgr   VaultManager
	txOutStore TxOutStore
}

// NewValidatorManager create a new instance of ValidatorManager
func NewValidatorMgr(k Keeper, txOut TxOutStore, vaultMgr VaultManager) *ValidatorMgr {
	return &ValidatorMgr{
		k:          k,
		vaultMgr:   vaultMgr,
		txOutStore: txOut,
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

	desireValidatorSet := constAccessor.GetInt64Value(constants.DesireValidatorSet)
	rotatePerBlockHeight := constAccessor.GetInt64Value(constants.RotatePerBlockHeight)
	if ctx.BlockHeight()%rotatePerBlockHeight == 0 {
		ctx.Logger().Info("Checking for node account rotation...")
		next, ok, err := vm.nextVaultNodeAccounts(ctx, int(desireValidatorSet))
		if err != nil {
			return err
		}
		if ok {
			if err := vm.vaultMgr.TriggerKeygen(ctx, next); err != nil {
				return err
			}
		}
	}

	return nil
}

// EndBlock when block end
func (vm *ValidatorMgr) EndBlock(ctx sdk.Context, constAccessor constants.ConstantValues) []abci.ValidatorUpdate {
	height := ctx.BlockHeight()
	activeNodes, err := vm.k.ListActiveNodeAccounts(ctx)
	if err != nil {
		ctx.Logger().Error("fail to list active node accounts")
	}

	readyNodes, err := vm.k.ListNodeAccountsByStatus(ctx, NodeReady)
	if err != nil {
		ctx.Logger().Error("fail to list ready node accounts")
	}

	active, err := vm.k.GetAsgardVaultsByStatus(ctx, ActiveVault)
	if err != nil {
		ctx.Logger().Error("fail to get active asgards")
	}

	// if we have no pool addresses, nothing to do...
	if len(active) == 0 {
		return nil
	}

	var membership common.PubKeys
	for _, vault := range active {
		membership = append(membership, vault.Membership...)
	}

	var newActive NodeAccounts    // store the list of new active users
	var removedNodes NodeAccounts // nodes that had been removed
	// find active node accounts that are no longer active

	for _, na := range activeNodes {
		found := false
		for _, vault := range active {
			if vault.Contains(na.PubKeySet.Secp256k1) {
				found = true
				newActive = append(newActive, na)
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
			removedNodes = append(removedNodes, na)
			if err := vm.k.SetNodeAccount(ctx, na); err != nil {
				ctx.Logger().Error("fail to save node account")
			}
		}
	}
	newNodesBecomeActive := false
	// find ready nodes that change to
	for _, na := range readyNodes {
		for _, member := range membership {
			if na.PubKeySet.Contains(member) {
				newActive = append(newActive, na)
				ctx.EventManager().EmitEvent(
					sdk.NewEvent("UpdateNodeAccountStatus",
						sdk.NewAttribute("Address", na.NodeAddress.String()),
						sdk.NewAttribute("Former:", na.Status.String()),
						sdk.NewAttribute("Current:", NodeActive.String())))
				newNodesBecomeActive = true
				na.UpdateStatus(NodeActive, height)
				if err := vm.k.SetNodeAccount(ctx, na); err != nil {
					ctx.Logger().Error("fail to save node account")
				}
				break
			}
		}
	}
	// no new nodes become active, and no nodes get removed , so
	if !newNodesBecomeActive && len(removedNodes) == 0 {
		return nil
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
	// if we remove a validator , we need to make sure their voting power get reset to 0
	for _, item := range removedNodes {
		pk, err := sdk.GetConsPubKeyBech32(item.ValidatorConsPubKey)
		if nil != err {
			ctx.Logger().Error("fail to parse consensus public key", "key", item.ValidatorConsPubKey)
			continue
		}
		validators = append(validators, abci.ValidatorUpdate{
			PubKey: tmtypes.TM2PB.PubKey(pk),
			Power:  0,
		})
	}

	minimumNodesForBFT := constAccessor.GetInt64Value(constants.MinimumNodesForBFT)
	if height > 1 && len(removedNodes) > 0 && len(newActive) < int(minimumNodesForBFT) { // THORNode still have enough validators for BFT
		if err := vm.processRagnarok(ctx, activeNodes, constAccessor); err != nil {
			ctx.Logger().Error("fail to process ragnarok protocol: %s", err)
		}
	}

	return validators
}

// determines when/if to run each part of the ragnarok process
func (vm *ValidatorMgr) processRagnarok(ctx sdk.Context, activeNodes NodeAccounts, constAccessor constants.ConstantValues) error {
	// execute Ragnarok protocol, no going back
	// THORNode have to request the fund back now, because once it get to the rotate block height ,
	// THORNode won't have validators anymore
	ragnarokHeight, err := vm.k.GetRagnarokBlockHeight(ctx)
	if err != nil {
		ctx.Logger().Error("fail to get ragnarok height", err)
		return err
	}

	if ragnarokHeight == 0 {
		ragnarokHeight = ctx.BlockHeight()
		vm.k.SetRagnarokBlockHeight(ctx, ragnarokHeight)
		if err := vm.ragnarokProtocolStage1(ctx, activeNodes); nil != err {
			ctx.Logger().Error(fmt.Errorf("fail to execute ragnarok protocol step 1: %w", err).Error())
			return err
		}
		return nil
	}

	migrateInterval := constAccessor.GetInt64Value(constants.FundMigrationInterval)
	if (ctx.BlockHeight()-ragnarokHeight)%migrateInterval == 0 {
		nth := (ctx.BlockHeight() - ragnarokHeight) / migrateInterval
		err := vm.ragnarokProtocolStage2(ctx, nth, constAccessor)
		if err != nil {
			ctx.Logger().Error("fail to execute ragnarok protocol step 2: %s", err)
			return err
		}
	}

	return nil
}

// ragnarokProtocolStage1 - request all yggdrasil pool to return the fund
// when THORNode observe the node return fund successfully, the node's bound will be refund.
func (vm *ValidatorMgr) ragnarokProtocolStage1(ctx sdk.Context, activeNodes NodeAccounts) error {
	return vm.recallYggFunds(ctx, activeNodes)
}

func (vm *ValidatorMgr) ragnarokProtocolStage2(ctx sdk.Context, nth int64, constAccessor constants.ConstantValues) error {
	// Ragnarok Protocol
	// If THORNode can no longer be BFT, do a graceful shutdown of the entire network.
	// 1) THORNode will request all yggdrasil pool to return fund , if THORNode don't have yggdrasil pool THORNode will go to step 3 directly
	// 2) upon receiving the yggdrasil fund,  THORNode will refund the validator's bond
	// 3) once all yggdrasil fund get returned, return all fund to stakes

	// refund bonders
	if err := vm.ragnarokBond(ctx, nth); err != nil {
		return err
	}

	// refund stakers
	if err := vm.ragnarokPools(ctx, nth, constAccessor); err != nil {
		return err
	}

	// refund reserve contributors
	if err := vm.ragnarokReserve(ctx, nth); err != nil {
		return err
	}

	return nil
}

func (vm *ValidatorMgr) ragnarokReserve(ctx sdk.Context, nth int64) error {
	contribs, err := vm.k.GetReservesContributors(ctx)
	if nil != err {
		ctx.Logger().Error("can't get reserve contributors", err)
		return err
	}
	vaultData, err := vm.k.GetVaultData(ctx)
	if nil != err {
		ctx.Logger().Error("can't get vault data", err)
		return err
	}
	totalReserve := vaultData.TotalReserve
	totalContributions := sdk.ZeroUint()
	for _, contrib := range contribs {
		totalContributions = totalContributions.Add(contrib.Amount)
	}

	// Since reserves are spent over time (via block rewards), reserve
	// contributors do not get back the full amounts they put in. Instead they
	// should get a percentage of the remaining amount, relative to the amount
	// they contributed. We'll be reducing the total reserve supply as we
	// refund reserves

	// nth * 10 == the amount of the bond we want to send
	for i, contrib := range contribs {
		share := common.GetShare(
			contrib.Amount,
			totalReserve,
			totalContributions,
		)
		if nth > 10 { // cap at 10
			nth = 10
		}
		amt := share.MulUint64(uint64(nth)).QuoUint64(10)
		vaultData.TotalReserve = common.SafeSub(vaultData.TotalReserve, amt)
		contribs[i].Amount = common.SafeSub(contrib.Amount, amt)

		// refund contribution
		txOutItem := &TxOutItem{
			Chain:     common.BNBChain,
			ToAddress: contrib.Address,
			InHash:    common.BlankTxID,
			Coin:      common.NewCoin(common.RuneAsset(), amt),
		}
		vm.txOutStore.AddTxOutItem(ctx, txOutItem)
	}

	if err := vm.k.SetVaultData(ctx, vaultData); err != nil {
		return err
	}

	if err := vm.k.SetReserveContributors(ctx, contribs); err != nil {
		return err
	}

	return nil
}

func (vm *ValidatorMgr) ragnarokBond(ctx sdk.Context, nth int64) error {
	active, err := vm.k.ListActiveNodeAccounts(ctx)
	if nil != err {
		ctx.Logger().Error("can't get active nodes", err)
		return err
	}

	// nth * 10 == the amount of the bond we want to send
	for _, na := range active {
		ygg, err := vm.k.GetVault(ctx, na.PubKeySet.Secp256k1)
		if err != nil {
			return err
		}
		if ygg.HasFunds() {
			ctx.Logger().Info(fmt.Sprintf("skip bond refund due to remaining funds: %s", na.NodeAddress))
			continue
		}
		if nth > 10 { // cap at 10
			nth = 10
		}
		amt := na.Bond.MulUint64(uint64(nth)).QuoUint64(10)
		na.Bond = common.SafeSub(na.Bond, amt)
		if err := vm.k.SetNodeAccount(ctx, na); err != nil {
			return err
		}

		// refund bond
		txOutItem := &TxOutItem{
			Chain:     common.BNBChain,
			ToAddress: na.BondAddress,
			InHash:    common.BlankTxID,
			Coin:      common.NewCoin(common.RuneAsset(), amt),
		}
		vm.txOutStore.AddTxOutItem(ctx, txOutItem)

	}

	return nil
}

func (vm *ValidatorMgr) ragnarokPools(ctx sdk.Context, nth int64, constAccessor constants.ConstantValues) error {
	nas, err := vm.k.ListActiveNodeAccounts(ctx)
	if nil != err {
		ctx.Logger().Error("can't get active nodes", err)
		return err
	}
	if len(nas) == 0 {
		return fmt.Errorf("can't find any active nodes")
	}
	na := nas[0]

	// each round of refund, we increase the percentage by 10%. This ensures
	// that we slowly refund each person, while not sending out too much too
	// fast. Also, we won't be running into any gas related issues until the
	// very last round, which, by my calculations, if someone staked 100 coins,
	// the last tx will send them 0.036288. So if we don't have enough gas to
	// send them, its only a very small portion that is not refunded.
	var basisPoints int64
	if nth > 10 {
		basisPoints = MaxWithdrawBasisPoints
	} else {
		basisPoints = nth * (MaxWithdrawBasisPoints / 10)
	}

	// go through all the pooles
	pools, err := vm.k.GetPools(ctx)
	if err != nil {
		ctx.Logger().Error("can't get pools", err)
		return err
	}
	for _, pool := range pools {
		poolStaker, err := vm.k.GetPoolStaker(ctx, pool.Asset)
		if nil != err {
			ctx.Logger().Error("fail to get pool staker", err)
			return err
		}

		// everyone withdraw
		for _, item := range poolStaker.Stakers {
			if item.Units.IsZero() {
				continue
			}

			unstakeMsg := NewMsgSetUnStake(
				common.GetRagnarokTx(pool.Asset.Chain),
				item.RuneAddress,
				sdk.NewUint(uint64(basisPoints)),
				pool.Asset,
				na.NodeAddress,
			)

			version := vm.k.GetLowestActiveVersion(ctx)
			unstakeHandler := NewUnstakeHandler(vm.k, vm.txOutStore)
			result := unstakeHandler.Run(ctx, unstakeMsg, version, constAccessor)
			if !result.IsOK() {
				ctx.Logger().Error("fail to unstake", "staker", item.RuneAddress)
				return fmt.Errorf("fail to unstake address: %s", result.Log)
			}
		}
		pool.Status = PoolBootstrap
		if err := vm.k.SetPool(ctx, pool); err != nil {
			ctx.Logger().Error(err.Error())
			return err
		}
	}

	return nil
}

func (vm *ValidatorMgr) RequestYggReturn(ctx sdk.Context, node NodeAccount) error {
	ygg, err := vm.k.GetVault(ctx, node.PubKeySet.Secp256k1)
	if nil != err {
		return fmt.Errorf("fail to get yggdrasil: %w", err)
	}
	if ygg.IsAsgard() {
		return nil
	}
	if !ygg.HasFunds() {
		return nil
	}

	chains, err := vm.k.GetChains(ctx)
	if err != nil {
		return err
	}

	active, err := vm.k.GetAsgardVaultsByStatus(ctx, ActiveVault)
	if err != nil {
		return err
	}

	vault := active.SelectByMinCoin(common.RuneAsset())
	if vault.IsEmpty() {
		return fmt.Errorf("unable to determine asgard vault")
	}

	for _, chain := range chains {
		toAddr, err := vault.PubKey.GetAddress(chain)
		if err != nil {
			return err
		}

		if !toAddr.IsEmpty() {
			txOutItem := &TxOutItem{
				Chain:       chain,
				ToAddress:   toAddr,
				InHash:      common.BlankTxID,
				VaultPubKey: ygg.PubKey,
				Memo:        "yggdrasil-",
			}
			vm.txOutStore.AddTxOutItem(ctx, txOutItem)
		}
	}

	return nil
}

func (vm *ValidatorMgr) recallYggFunds(ctx sdk.Context, activeNodes NodeAccounts) error {
	// request every node to return fund
	for _, na := range activeNodes {
		if err := vm.RequestYggReturn(ctx, na); nil != err {
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
	if !na.IsEmpty() && na.LeaveHeight == 0 {
		ctx.Logger().Info(fmt.Sprintf("Marked Validator to be churned out %s", na.NodeAddress))
		na.LeaveHeight = ctx.BlockHeight()
		return vm.k.SetNodeAccount(ctx, na)
	}
	return nil
}

// Mark an old actor to be churned out
func (vm *ValidatorMgr) markOldActor(ctx sdk.Context, rate int64) error {
	if ctx.BlockHeight()%rate == 0 {
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
	if ctx.BlockHeight()%rate == 0 {
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
		na.UpdateStatus(NodeReady, ctx.BlockHeight()) // everyone starts with the benefit of the doubt
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
func (vm *ValidatorMgr) nextVaultNodeAccounts(ctx sdk.Context, targetCount int) (NodeAccounts, bool, error) {
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
		return active[i].LeaveHeight > active[j].LeaveHeight
	})

	// remove a node node account, if one is marked to leave
	if len(active) > 0 && (active[0].LeaveHeight > 0 || active[0].RequestedToLeave) {
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
