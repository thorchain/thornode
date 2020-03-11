package thorchain

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

// validatorMgrV1 is to manage a list of validators , and rotate them
type validatorMgrV1 struct {
	version               semver.Version
	k                     Keeper
	versionedVaultManager VersionedVaultManager
	versionedTxOutStore   VersionedTxOutStore
}

// newValidatorMgrV1 create a new instance of ValidatorManager
func newValidatorMgrV1(k Keeper, versionedTxOutStore VersionedTxOutStore, versionedVaultManager VersionedVaultManager) *validatorMgrV1 {
	return &validatorMgrV1{
		version:               semver.MustParse("0.1.0"),
		k:                     k,
		versionedVaultManager: versionedVaultManager,
		versionedTxOutStore:   versionedTxOutStore,
	}
}

// BeginBlock when block begin
func (vm *validatorMgrV1) BeginBlock(ctx sdk.Context, constAccessor constants.ConstantValues) error {
	height := ctx.BlockHeight()
	if height == genesisBlockHeight {
		if err := vm.setupValidatorNodes(ctx, height, constAccessor); err != nil {
			ctx.Logger().Error("fail to setup validator nodes", "error", err)
		}
	}
	if vm.k.RagnarokInProgress(ctx) {
		// ragnarok is in progress, no point to check node rotation
		return nil
	}
	vaultMgr, err := vm.versionedVaultManager.GetVaultManager(ctx, vm.k, vm.version)
	if err != nil {
		return fmt.Errorf("fail to get a valid vault: %w", err)
	}
	minimumNodesForBFT := constAccessor.GetInt64Value(constants.MinimumNodesForBFT)
	totalActiveNodes, err := vm.k.TotalActiveNodeAccount(ctx)
	if err != nil {
		return err
	}

	artificialRagnarokBlockHeight := constAccessor.GetInt64Value(constants.ArtificialRagnarokBlockHeight)
	if minimumNodesForBFT+2 < int64(totalActiveNodes) ||
		(artificialRagnarokBlockHeight > 0 && ctx.BlockHeight() >= artificialRagnarokBlockHeight) {
		badValidatorRate := constAccessor.GetInt64Value(constants.BadValidatorRate)
		if err := vm.markBadActor(ctx, badValidatorRate); err != nil {
			return err
		}
		oldValidatorRate := constAccessor.GetInt64Value(constants.OldValidatorRate)
		if err := vm.markOldActor(ctx, oldValidatorRate); err != nil {
			return err
		}
	}

	// calculate last churn block height
	vaults, err := vm.k.GetAsgardVaultsByStatus(ctx, ActiveVault)
	if err != nil {
		return err
	}
	var lastHeight int64 // the last block height we had a successful churn
	for _, vault := range vaults {
		if vault.BlockHeight > lastHeight {
			lastHeight = vault.BlockHeight
		}
	}

	// get constants
	desireValidatorSet := constAccessor.GetInt64Value(constants.DesireValidatorSet)
	rotatePerBlockHeight := constAccessor.GetInt64Value(constants.RotatePerBlockHeight)
	rotateRetryBlocks := constAccessor.GetInt64Value(constants.RotateRetryBlocks)

	// calculate if we need to retry a churn because we are overdue for a
	// successful one
	retryChurn := ctx.BlockHeight()-lastHeight > rotatePerBlockHeight && (ctx.BlockHeight()-lastHeight+rotatePerBlockHeight)%rotateRetryBlocks == 0

	if ctx.BlockHeight()%rotatePerBlockHeight == 0 || retryChurn {
		if retryChurn {
			ctx.Logger().Info("Checking for node account rotation... (retry)")
		} else {
			ctx.Logger().Info("Checking for node account rotation...")
		}

		// don't churn if we have retiring asgard vaults that still have funds
		retiringVaults, err := vm.k.GetAsgardVaultsByStatus(ctx, RetiringVault)
		if err != nil {
			return err
		}
		for _, vault := range retiringVaults {
			if vault.HasFunds() {
				ctx.Logger().Info("Skipping rotation due to retiring vaults still have funds.")
				return nil
			}
		}

		next, ok, err := vm.nextVaultNodeAccounts(ctx, int(desireValidatorSet), constAccessor)
		if err != nil {
			return err
		}
		if ok {
			if err := vaultMgr.TriggerKeygen(ctx, next); err != nil {
				return err
			}
		}
	}

	return nil
}

// EndBlock when block end
func (vm *validatorMgrV1) EndBlock(ctx sdk.Context, constAccessor constants.ConstantValues) []abci.ValidatorUpdate {
	height := ctx.BlockHeight()
	activeNodes, err := vm.k.ListActiveNodeAccounts(ctx)
	if err != nil {
		ctx.Logger().Error("fail to get all active nodes", "error", err)
		return nil
	}

	// when ragnarok is in progress, just process ragnarok
	if vm.k.RagnarokInProgress(ctx) {
		// process ragnarok
		if err := vm.processRagnarok(ctx, activeNodes, constAccessor); err != nil {
			ctx.Logger().Error("fail to process ragnarok protocol", "error", err)
		}
		return nil
	}

	newNodes, removedNodes, err := vm.getChangedNodes(ctx, activeNodes)
	if err != nil {
		ctx.Logger().Error("fail to get node changes", "error", err)
		return nil
	}

	// no change
	if len(newNodes) == 0 && len(removedNodes) == 0 {
		return nil
	}

	minimumNodesForBFT := constAccessor.GetInt64Value(constants.MinimumNodesForBFT)
	nodesAfterChange := len(activeNodes) + len(newNodes) - len(removedNodes)
	if len(activeNodes) >= int(minimumNodesForBFT) && nodesAfterChange < int(minimumNodesForBFT) {
		// THORNode don't have enough validators for BFT
		if err := vm.processRagnarok(ctx, activeNodes, constAccessor); err != nil {
			ctx.Logger().Error("fail to process ragnarok protocol", "error", err)
		}
		// by return
		return nil
	}
	validators := make([]abci.ValidatorUpdate, 0, len(newNodes)+len(removedNodes))
	for _, na := range newNodes {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent("UpdateNodeAccountStatus",
				sdk.NewAttribute("Address", na.NodeAddress.String()),
				sdk.NewAttribute("Former:", na.Status.String()),
				sdk.NewAttribute("Current:", NodeActive.String())))
		na.UpdateStatus(NodeActive, height)
		if err := vm.k.SetNodeAccount(ctx, na); err != nil {
			ctx.Logger().Error("fail to save node account", "error", err)
		}
		pk, err := sdk.GetConsPubKeyBech32(na.ValidatorConsPubKey)
		if err != nil {
			ctx.Logger().Error("fail to parse consensus public key", "key", na.ValidatorConsPubKey, "error", err)
			continue
		}
		validators = append(validators, abci.ValidatorUpdate{
			PubKey: tmtypes.TM2PB.PubKey(pk),
			Power:  100,
		})
	}
	for _, na := range removedNodes {
		status := NodeStandby
		if na.RequestedToLeave {
			status = NodeDisabled
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent("UpdateNodeAccountStatus",
				sdk.NewAttribute("Address", na.NodeAddress.String()),
				sdk.NewAttribute("Former:", na.Status.String()),
				sdk.NewAttribute("Current:", status.String())))
		na.UpdateStatus(status, height)
		removedNodes = append(removedNodes, na)
		if err := vm.k.SetNodeAccount(ctx, na); err != nil {
			ctx.Logger().Error("fail to save node account", "error", err)
		}

		if err := vm.payNodeAccountBondAward(ctx, na); err != nil {
			ctx.Logger().Error("fail to pay node account bond award", "error", err)
		}
		pk, err := sdk.GetConsPubKeyBech32(na.ValidatorConsPubKey)
		if err != nil {
			ctx.Logger().Error("fail to parse consensus public key", "key", na.ValidatorConsPubKey, "error", err)
			continue
		}
		validators = append(validators, abci.ValidatorUpdate{
			PubKey: tmtypes.TM2PB.PubKey(pk),
			Power:  0,
		})
	}

	return validators
}

// getChangedNodes to identify which node had been removed ,and which one had been added
// newNodes , removed nodes,err
func (vm *validatorMgrV1) getChangedNodes(ctx sdk.Context, activeNodes NodeAccounts) (NodeAccounts, NodeAccounts, error) {
	var newActive NodeAccounts    // store the list of new active users
	var removedNodes NodeAccounts // nodes that had been removed

	readyNodes, err := vm.k.ListNodeAccountsByStatus(ctx, NodeReady)
	if err != nil {
		return newActive, removedNodes, fmt.Errorf("fail to list ready node accounts: %w", err)
	}

	active, err := vm.k.GetAsgardVaultsByStatus(ctx, ActiveVault)
	if err != nil {
		ctx.Logger().Error("fail to get active asgards", "error", err)
		return newActive, removedNodes, fmt.Errorf("fail to get active asgards: %w", err)
	}
	if len(active) == 0 {
		return newActive, removedNodes, errors.New("no active vault")
	}
	var membership common.PubKeys
	for _, vault := range active {
		membership = append(membership, vault.Membership...)
	}

	// find active node accounts that are no longer active
	for _, na := range activeNodes {
		found := false
		for _, vault := range active {
			if vault.Contains(na.PubKeySet.Secp256k1) {
				found = true
				break
			}
		}
		if !found && len(membership) > 0 {
			removedNodes = append(removedNodes, na)
		}
	}

	// find ready nodes that change to
	for _, na := range readyNodes {
		for _, member := range membership {
			if na.PubKeySet.Contains(member) {
				newActive = append(newActive, na)
				break
			}
		}
	}

	return newActive, removedNodes, nil
}

// payNodeAccountBondAward pay
func (vm *validatorMgrV1) payNodeAccountBondAward(ctx sdk.Context, na NodeAccount) error {
	if na.ActiveBlockHeight == 0 || na.Bond.IsZero() {
		return nil
	}
	// The node account seems to have become a non active node account.
	// Therefore, lets give them their bond rewards.
	vault, err := vm.k.GetVaultData(ctx)
	if err != nil {
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

	if err := vm.k.SetVaultData(ctx, vault); err != nil {
		return fmt.Errorf("fail to save vault data: %w", err)
	}
	na.ActiveBlockHeight = 0
	return vm.k.SetNodeAccount(ctx, na)
}

// determines when/if to run each part of the ragnarok process
func (vm *validatorMgrV1) processRagnarok(ctx sdk.Context, activeNodes NodeAccounts, constAccessor constants.ConstantValues) error {
	// execute Ragnarok protocol, no going back
	// THORNode have to request the fund back now, because once it get to the rotate block height ,
	// THORNode won't have validators anymore
	ragnarokHeight, err := vm.k.GetRagnarokBlockHeight(ctx)
	if err != nil {
		return fmt.Errorf("fail to get ragnarok height: %w", err)
	}

	if ragnarokHeight == 0 {
		ragnarokHeight = ctx.BlockHeight()
		vm.k.SetRagnarokBlockHeight(ctx, ragnarokHeight)
		if err := vm.ragnarokProtocolStage1(ctx, activeNodes); err != nil {
			return fmt.Errorf("fail to execute ragnarok protocol step 1: %w", err)
		}
		if err := vm.ragnarokBondReward(ctx); err != nil {
			return fmt.Errorf("when ragnarok triggered ,fail to give all active node bond reward %w", err)
		}
		return nil
	}

	migrateInterval := constAccessor.GetInt64Value(constants.FundMigrationInterval)
	if (ctx.BlockHeight()-ragnarokHeight)%migrateInterval == 0 {
		nth := (ctx.BlockHeight() - ragnarokHeight) / migrateInterval
		err := vm.ragnarokProtocolStage2(ctx, nth, constAccessor)
		if err != nil {
			ctx.Logger().Error("fail to execute ragnarok protocol step 2", "error", err)
			return err
		}
	}

	return nil
}

// ragnarokProtocolStage1 - request all yggdrasil pool to return the fund
// when THORNode observe the node return fund successfully, the node's bound will be refund.
func (vm *validatorMgrV1) ragnarokProtocolStage1(ctx sdk.Context, activeNodes NodeAccounts) error {
	return vm.recallYggFunds(ctx, activeNodes)
}

func (vm *validatorMgrV1) ragnarokProtocolStage2(ctx sdk.Context, nth int64, constAccessor constants.ConstantValues) error {
	// Ragnarok Protocol
	// If THORNode can no longer be BFT, do a graceful shutdown of the entire network.
	// 1) THORNode will request all yggdrasil pool to return fund , if THORNode don't have yggdrasil pool THORNode will go to step 3 directly
	// 2) upon receiving the yggdrasil fund,  THORNode will refund the validator's bond
	// 3) once all yggdrasil fund get returned, return all fund to stakes

	// refund bonders
	if err := vm.ragnarokBond(ctx, nth); err != nil {
		return err
	}

	// refund reserve contributors
	if err := vm.ragnarokReserve(ctx, nth); err != nil {
		return err
	}

	// refund stakers. This is last to ensure there is likely gas for the
	// returning bond and reserve
	if err := vm.ragnarokPools(ctx, nth, constAccessor); err != nil {
		return err
	}

	return nil
}

func (vm *validatorMgrV1) ragnarokBondReward(ctx sdk.Context) error {
	active, err := vm.k.ListActiveNodeAccounts(ctx)
	if err != nil {
		return fmt.Errorf("fail to get all active node account: %w", err)
	}
	for _, item := range active {
		if err := vm.payNodeAccountBondAward(ctx, item); err != nil {
			return fmt.Errorf("fail to pay node account(%s) bond award: %w", item.NodeAddress.String(), err)
		}
	}
	return nil
}

func (vm *validatorMgrV1) ragnarokReserve(ctx sdk.Context, nth int64) error {
	contribs, err := vm.k.GetReservesContributors(ctx)
	if err != nil {
		ctx.Logger().Error("can't get reserve contributors", "error", err)
		return err
	}
	if len(contribs) == 0 {
		return nil
	}
	vaultData, err := vm.k.GetVaultData(ctx)
	if err != nil {
		ctx.Logger().Error("can't get vault data", "error", err)
		return err
	}

	if vaultData.TotalReserve.IsZero() {
		return nil
	}

	txOutStore, err := vm.versionedTxOutStore.GetTxOutStore(vm.k, vm.version)
	if err != nil {
		ctx.Logger().Error("can't get tx out store", "error", err)
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
			Memo:      NewRagnarokMemo(ctx.BlockHeight()).String(),
		}
		_, err = txOutStore.TryAddTxOutItem(ctx, txOutItem)
		if err != nil {
			return fmt.Errorf("fail to add outbound transaction")
		}
	}

	if err := vm.k.SetVaultData(ctx, vaultData); err != nil {
		return err
	}

	if err := vm.k.SetReserveContributors(ctx, contribs); err != nil {
		return err
	}

	return nil
}

func (vm *validatorMgrV1) ragnarokBond(ctx sdk.Context, nth int64) error {
	nas, err := vm.k.ListNodeAccounts(ctx)
	if err != nil {
		ctx.Logger().Error("can't get nodes", "error", err)
		return err
	}
	txOutStore, err := vm.versionedTxOutStore.GetTxOutStore(vm.k, vm.version)
	if err != nil {
		ctx.Logger().Error("can't get tx out store", "error", err)
		return err
	}
	// nth * 10 == the amount of the bond we want to send
	for _, na := range nas {
		if na.Bond.IsZero() {
			continue
		}
		if vm.k.VaultExists(ctx, na.PubKeySet.Secp256k1) {
			ygg, err := vm.k.GetVault(ctx, na.PubKeySet.Secp256k1)
			if err != nil {
				return err
			}
			if ygg.HasFunds() {
				ctx.Logger().Info(fmt.Sprintf("skip bond refund due to remaining funds: %s", na.NodeAddress))
				continue
			}
		}

		if nth > 10 { // cap at 10
			nth = 10
		}
		amt := na.Bond.MulUint64(uint64(nth)).QuoUint64(10)

		// refund bond
		txOutItem := &TxOutItem{
			Chain:     common.BNBChain,
			ToAddress: na.BondAddress,
			InHash:    common.BlankTxID,
			Coin:      common.NewCoin(common.RuneAsset(), amt),
			Memo:      NewRagnarokMemo(ctx.BlockHeight()).String(),
		}
		ok, err := txOutStore.TryAddTxOutItem(ctx, txOutItem)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		na.Bond = common.SafeSub(na.Bond, amt)
		if err := vm.k.SetNodeAccount(ctx, na); err != nil {
			return err
		}
	}

	return nil
}

func (vm *validatorMgrV1) ragnarokPools(ctx sdk.Context, nth int64, constAccessor constants.ConstantValues) error {
	nas, err := vm.k.ListActiveNodeAccounts(ctx)
	if err != nil {
		ctx.Logger().Error("can't get active nodes", "error", err)
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
		basisPoints = MaxUnstakeBasisPoints
	} else {
		basisPoints = nth * (MaxUnstakeBasisPoints / 10)
	}

	// go through all the pools
	pools, err := vm.k.GetPools(ctx)
	if err != nil {
		ctx.Logger().Error("can't get pools", "error", err)
		return err
	}

	// set all pools to bootstrap mode
	for _, pool := range pools {
		pool.Status = PoolBootstrap
		if err := vm.k.SetPool(ctx, pool); err != nil {
			ctx.Logger().Error(err.Error())
			return err
		}
	}

	for i := len(pools) - 1; i >= 0; i-- { // iterate backwards
		pool := pools[i]
		poolStaker, err := vm.k.GetPoolStaker(ctx, pool.Asset)
		if err != nil {
			ctx.Logger().Error("fail to get pool staker", "error", err)
			return err
		}

		// everyone withdraw
		for i := len(poolStaker.Stakers) - 1; i >= 0; i-- { // iterate backwards
			item := poolStaker.Stakers[i]
			if item.Units.IsZero() {
				continue
			}

			unstakeMsg := NewMsgSetUnStake(
				common.GetRagnarokTx(pool.Asset.Chain, item.RuneAddress, item.RuneAddress),
				item.RuneAddress,
				sdk.NewUint(uint64(basisPoints)),
				pool.Asset,
				na.NodeAddress,
			)

			version := vm.k.GetLowestActiveVersion(ctx)
			unstakeHandler := NewUnstakeHandler(vm.k, vm.versionedTxOutStore)
			result := unstakeHandler.Run(ctx, unstakeMsg, version, constAccessor)
			if !result.IsOK() {
				ctx.Logger().Error("fail to unstake", "staker", item.RuneAddress, "error", result.Log)
				return fmt.Errorf("fail to unstake address: %s", result.Log)
			}
		}
	}

	return nil
}

func (vm *validatorMgrV1) RequestYggReturn(ctx sdk.Context, node NodeAccount) error {
	if !vm.k.VaultExists(ctx, node.PubKeySet.Secp256k1) {
		return nil
	}
	ygg, err := vm.k.GetVault(ctx, node.PubKeySet.Secp256k1)
	if err != nil {
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
	txOutStore, err := vm.versionedTxOutStore.GetTxOutStore(vm.k, vm.version)
	if err != nil {
		ctx.Logger().Error("can't get tx out store", "error", err)
		return err
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
			// yggdrasil- will not set coin field here, when signer see a TxOutItem that has memo "yggdrasil-" it will query the chain
			// and find out all the remaining assets , and fill in the field
			_, err := txOutStore.TryAddTxOutItem(ctx, txOutItem)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *validatorMgrV1) recallYggFunds(ctx sdk.Context, activeNodes NodeAccounts) error {
	// request every node to return fund
	for _, na := range activeNodes {
		if err := vm.RequestYggReturn(ctx, na); err != nil {
			return fmt.Errorf("fail to request yggdrasil fund back: %w", err)
		}
	}
	return nil
}

// setupValidatorNodes it is one off it only get called when genesis
func (vm *validatorMgrV1) setupValidatorNodes(ctx sdk.Context, height int64, constAccessor constants.ConstantValues) error {
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
		if err := vm.k.Cdc().UnmarshalBinaryBare(iter.Value(), &na); err != nil {
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
		if err := vm.k.SetNodeAccount(ctx, item); err != nil {
			return fmt.Errorf("fail to save node account: %w", err)
		}
	}
	return nil
}

// Iterate over active node accounts, finding bad actors with high slash points
func (vm *validatorMgrV1) findBadActors(ctx sdk.Context) (NodeAccounts, error) {
	badActors := NodeAccounts{}
	nas, err := vm.k.ListActiveNodeAccounts(ctx)
	if err != nil {
		return badActors, err
	}

	if len(nas) == 0 {
		return nil, nil
	}

	// NOTE: Our score gives a numerical representation of the behavior our a
	// node account. The lower the score, the worse behavior. The score is
	// determined by relative to how many slash points they have over how long
	// they have been an active node account.
	type badTracker struct {
		Score       sdk.Dec
		NodeAccount NodeAccount
	}
	var tracker []badTracker
	totalScore := sdk.ZeroDec()

	// Find bad actor relative to age / slashpoints
	for _, na := range nas {
		if na.SlashPoints == 0 {
			continue
		}

		age := sdk.NewDecWithPrec(ctx.BlockHeight()-na.StatusSince, 5)
		if age.LT(sdk.NewDecWithPrec(720, 5)) {
			// this node account is too new (1 hour) to be considered for removal
			continue
		}
		score := age.Quo(sdk.NewDecWithPrec(na.SlashPoints, 5))
		totalScore = totalScore.Add(score)

		tracker = append(tracker, badTracker{
			Score:       score,
			NodeAccount: na,
		})
	}

	if len(tracker) == 0 {
		// no offenders, exit nicely
		return nil, nil
	}

	sort.SliceStable(tracker, func(i, j int) bool {
		return tracker[i].Score.LT(tracker[j].Score)
	})

	avgScore := totalScore.QuoInt64(int64(len(nas)))

	// NOTE: our redline is a hard line in the sand to determine if a node
	// account is sufficiently bad that it should just be removed now. This
	// ensures that if we have multiple "really bad" node accounts, they all
	// can get removed in the same churn. It is important to note we shouldn't
	// be able to churn out more than 1/3rd of our node accounts in a single
	// churn, as that could threaten the security of the funds. This logic to
	// protect against this is not inside this function.
	redline := avgScore.QuoInt64(3)

	// find any node accounts that have crossed the redline
	for _, track := range tracker {
		if redline.GTE(track.Score) {
			badActors = append(badActors, track.NodeAccount)
		}
	}

	// if no one crossed the redline, lets just grab the worse offender
	if len(badActors) == 0 {
		badActors = NodeAccounts{tracker[0].NodeAccount}
	}

	return badActors, nil
}

// Iterate over active node accounts, finding the one that has been active longest
func (vm *validatorMgrV1) findOldActor(ctx sdk.Context) (NodeAccount, error) {
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
func (vm *validatorMgrV1) markActor(ctx sdk.Context, na NodeAccount) error {
	if !na.IsEmpty() && na.LeaveHeight == 0 {
		ctx.Logger().Info(fmt.Sprintf("Marked Validator to be churned out %s", na.NodeAddress))
		na.LeaveHeight = ctx.BlockHeight()
		return vm.k.SetNodeAccount(ctx, na)
	}
	return nil
}

// Mark an old actor to be churned out
func (vm *validatorMgrV1) markOldActor(ctx sdk.Context, rate int64) error {
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
func (vm *validatorMgrV1) markBadActor(ctx sdk.Context, rate int64) error {
	if ctx.BlockHeight()%rate == 0 {
		nas, err := vm.findBadActors(ctx)
		if err != nil {
			return err
		}
		for _, na := range nas {
			if err := vm.markActor(ctx, na); err != nil {
				return err
			}
		}
	}
	return nil
}

// find any actor that are ready to become "ready" status
func (vm *validatorMgrV1) markReadyActors(ctx sdk.Context, constAccessor constants.ConstantValues) error {
	standby, err := vm.k.ListNodeAccountsByStatus(ctx, NodeStandby)
	if err != nil {
		return err
	}
	ready, err := vm.k.ListNodeAccountsByStatus(ctx, NodeReady)
	if err != nil {
		return err
	}
	artificialRagnarokBlockHeight := constAccessor.GetInt64Value(constants.ArtificialRagnarokBlockHeight)
	if artificialRagnarokBlockHeight > 0 && ctx.BlockHeight() >= artificialRagnarokBlockHeight {
		// ArtificialRagnarokBlockHeight should only have a positive value on
		// chaosnet , we could even remove this after chaosnet

		// mark every node to standby, thus no node will be rotated in.
		for _, na := range append(standby, ready...) {
			na.UpdateStatus(NodeStandby, ctx.BlockHeight())
			if err := vm.k.SetNodeAccount(ctx, na); err != nil {
				return err
			}
		}
		return nil
	}

	// find min version node has to be, to be "ready" status
	minVersion := vm.k.GetMinJoinVersion(ctx)

	// check all ready and standby nodes are in "ready" state (upgrade/downgrade as needed)
	for _, na := range append(standby, ready...) {
		na.UpdateStatus(NodeReady, ctx.BlockHeight()) // everyone starts with the benefit of the doubt
		// TODO: check node is up to date on binance, etc
		// must have made an observation that matched 2/3rds within the last 5 blocks
		// We do not need to check thorchain is up to date because keygen will
		// fail if they are not.

		// Check version number is still supported
		if na.Version.LT(minVersion) {
			na.UpdateStatus(NodeStandby, ctx.BlockHeight())
		}

		if na.RequestedToLeave {
			na.UpdateStatus(NodeStandby, ctx.BlockHeight())
		}

		// Check if they've requested to leave
		if na.RequestedToLeave {
			na.UpdateStatus(NodeStandby, ctx.BlockHeight())
		}

		if err := vm.k.SetNodeAccount(ctx, na); err != nil {
			return err
		}
	}

	return nil
}

// Returns a list of nodes to include in the next pool
func (vm *validatorMgrV1) nextVaultNodeAccounts(ctx sdk.Context, targetCount int, constAccessor constants.ConstantValues) (NodeAccounts, bool, error) {
	rotation := false // track if are making any changes to the current active node accounts

	// update list of ready actors
	if err := vm.markReadyActors(ctx, constAccessor); err != nil {
		return nil, false, err
	}

	ready, err := vm.k.ListNodeAccountsByStatus(ctx, NodeReady)
	if err != nil {
		return nil, false, err
	}

	ready = sortNodeAccountsByProbabilisticBond(ctx, ready)

	active, err := vm.k.ListActiveNodeAccounts(ctx)
	if err != nil {
		return nil, false, err
	}
	// sort by LeaveHeight, giving preferential treatment to people who
	// requested to leave
	sort.SliceStable(active, func(i, j int) bool {
		if active[i].RequestedToLeave != active[j].RequestedToLeave {
			return active[i].RequestedToLeave
		}
		return active[i].LeaveHeight > active[j].LeaveHeight
	})

	artificialRagnarokBlockHeight := constAccessor.GetInt64Value(constants.ArtificialRagnarokBlockHeight)
	toRemove := findCounToRemove(ctx.BlockHeight(), artificialRagnarokBlockHeight, active)
	if toRemove > 0 {
		rotation = true
		active = active[toRemove:]
	}

	// add ready nodes to become active
	limit := toRemove + 1 // Max limit of ready nodes to churn in
	for i := 1; i < targetCount-len(active); i++ {
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

// findCounToRemove - find the number of node accounts to remove
func findCounToRemove(blockHeight, artificalRagnarok int64, active NodeAccounts) (toRemove int) {
	// count number of node accounts that are a candidate to leaving
	var candidateCount int
	for _, na := range active {
		if na.LeaveHeight > 0 {
			candidateCount += 1
			continue
		}
		break
	}

	maxRemove := findMaxAbleToLeave(len(active))
	if len(active) > 0 {
		if maxRemove == 0 {
			// we can't remove any mathematically, but we always leave room for node
			// accounts requesting to leave
			if active[0].RequestedToLeave || (artificalRagnarok > 0 && blockHeight >= artificalRagnarok) {
				toRemove = 1
			}
		} else {
			if candidateCount > maxRemove {
				toRemove = maxRemove
			} else {
				toRemove = candidateCount
			}
		}
	}
	return
}

// findMaxAbleToLeave - given number of current active node account, figure out
// the max number of individuals we can allow to leave in a single churn event
func findMaxAbleToLeave(count int) int {
	majority := (count * 2 / 3) + 1 // add an extra 1 to "round up" for security
	max := count - majority

	// we don't want to loose BFT by accident (only when someone leaves)
	if count-max < 4 {
		max = count - 4
		if max < 0 {
			max = 0
		}
	}

	return max
}

// sortNodeAccountsByProbabilisticBond - this takes a list of node accounts and
// sorts them in a deterministic random order with relationship to the bond
// amounts.
func sortNodeAccountsByProbabilisticBond(ctx sdk.Context, nas NodeAccounts) (result NodeAccounts) {
	// get seed random number from ctx.TxBytes()
	// we use TxBytes because its not a number that someone can reasonably
	// be predicted for a future block
	var seed int64
	for _, x := range ctx.TxBytes() {
		seed += int64(x)
	}
	source := rand.NewSource(seed)
	rnd := rand.New(source)

	// sort by node address
	sort.SliceStable(nas, func(i, j int) bool {
		return nas[i].NodeAddress.String() > nas[j].NodeAddress.String()
	})

	// we sort our list of nodes with a relationship to the amount of
	// bond they have bonded. Higher bond should give them a higher chance of
	// be selected.
	for len(nas) > 0 {
		nextReadyList := make(NodeAccounts, 0) // store our list of non-selected node accounts

		// sum bond of nodes
		totalBond := sdk.ZeroUint()
		for _, na := range nas {
			totalBond = totalBond.Add(na.Bond)
		}

		if totalBond.IsZero() {
			return nas
		}

		// get our randomly chosen number within our total bond
		idx := rnd.Uint64() % totalBond.Uint64()

		var count uint64
		for _, na := range nas {
			// if our idx falls inside the bond of this node, they are the next
			// chosen node accounts.
			if (count+na.Bond.Uint64()) >= idx && count < idx {
				result = append(result, na)
			} else {
				nextReadyList = append(nextReadyList, na)
			}
			count += na.Bond.Uint64()
		}

		nas = nextReadyList
	}

	return
}
