package thorchain

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
	q "gitlab.com/thorchain/thornode/x/thorchain/query"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper, validatorMgr VersionedValidatorManager) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case q.QueryPool.Key:
			return queryPool(ctx, path[1:], req, keeper)
		case q.QueryPools.Key:
			return queryPools(ctx, req, keeper)
		case q.QueryStakers.Key:
			return queryStakers(ctx, path[1:], req, keeper)
		case q.QueryTxIn.Key:
			return queryTxIn(ctx, path[1:], req, keeper)
		case q.QueryKeysignArray.Key:
			return queryKeysign(ctx, path[1:], req, keeper)
		case q.QueryKeysignArrayPubkey.Key:
			return queryKeysign(ctx, path[1:], req, keeper)
		case q.QueryKeygensPubkey.Key:
			return queryKeygen(ctx, path[1:], req, keeper)
		case q.QueryCompEvents.Key:
			return queryCompEvents(ctx, path[1:], req, keeper)
		case q.QueryCompEventsByChain.Key:
			return queryCompEvents(ctx, path[1:], req, keeper)
		case q.QueryEventsByTxHash.Key:
			return queryEventsByTxHash(ctx, path[1:], req, keeper)
		case q.QueryHeights.Key:
			return queryHeights(ctx, path[1:], req, keeper)
		case q.QueryChainHeights.Key:
			return queryHeights(ctx, path[1:], req, keeper)
		case q.QueryObservers.Key:
			return queryObservers(ctx, path[1:], req, keeper)
		case q.QueryObserver.Key:
			return queryObserver(ctx, path[1:], req, keeper)
		case q.QueryNodeAccount.Key:
			return queryNodeAccount(ctx, path[1:], req, keeper)
		case q.QueryNodeAccounts.Key:
			return queryNodeAccounts(ctx, path[1:], req, keeper)
		case q.QueryPoolAddresses.Key:
			return queryPoolAddresses(ctx, path[1:], req, keeper)
		case q.QueryVaultData.Key:
			return queryVaultData(ctx, keeper)
		case q.QueryVaultsAsgard.Key:
			return queryAsgardVaults(ctx, keeper)
		case q.QueryVaultsYggdrasil.Key:
			return queryYggdrasilVaults(ctx, keeper)
		case q.QueryVaultPubkeys.Key:
			return queryVaultsPubkeys(ctx, keeper)
		case q.QueryTSSSigners.Key:
			return queryTSSSigners(ctx, path[1:], req, keeper)
		case q.QueryConstantValues.Key:
			return queryConstantValues(ctx, path[1:], req, keeper)
		case q.QueryMimirValues.Key:
			return queryMimirValues(ctx, path[1:], req, keeper)
		case q.QueryBan.Key:
			return queryBan(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest(
				fmt.Sprintf("unknown thorchain query endpoint: %s", path[0]),
			)
		}
	}
}

func getURLFromData(data []byte) (*url.URL, error) {
	if data == nil {
		return nil, errors.New("empty data")
	}
	u := &url.URL{}
	err := u.UnmarshalBinary(data)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal url.URL: %w", err)
	}
	return u, nil
}

func queryAsgardVaults(ctx sdk.Context, keeper Keeper) ([]byte, sdk.Error) {
	vaults, err := keeper.GetAsgardVaults(ctx)
	if err != nil {
		ctx.Logger().Error("fail to get asgard vaults", "error", err)
		return nil, sdk.ErrInternal("fail to get asgard vaults")
	}

	var vaultsWithFunds Vaults
	for _, vault := range vaults {
		if vault.IsAsgard() && (vault.HasFunds() || vault.Status == ActiveVault) {
			vaultsWithFunds = append(vaultsWithFunds, vault)
		}
	}

	res, err := codec.MarshalJSONIndent(keeper.Cdc(), vaultsWithFunds)
	if err != nil {
		ctx.Logger().Error("fail to marshal vaults response to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal response to json")
	}

	return res, nil
}

func queryYggdrasilVaults(ctx sdk.Context, keeper Keeper) ([]byte, sdk.Error) {
	vaults := make(Vaults, 0)
	iter := keeper.GetVaultIterator(ctx)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var vault Vault
		if err := keeper.Cdc().UnmarshalBinaryBare(iter.Value(), &vault); err != nil {
			ctx.Logger().Error("fail to unmarshal yggdrasil", "error", err)
			return nil, sdk.ErrInternal("fail to unmarshal yggdrasil")
		}
		if vault.IsYggdrasil() && vault.HasFunds() {
			vaults = append(vaults, vault)
		}
	}

	respVaults := make([]QueryYggdrasilVaults, len(vaults))
	for i, vault := range vaults {
		totalValue := sdk.ZeroUint()

		// find the bond of this node account
		na, err := keeper.GetNodeAccountByPubKey(ctx, vault.PubKey)
		if err != nil {
			ctx.Logger().Error("fail to get node account by pubkey", "error", err)
			continue
		}

		// calculate the total value of this yggdrasil vault
		for _, coin := range vault.Coins {
			if coin.Asset.IsRune() {
				totalValue = totalValue.Add(coin.Amount)
			} else {
				pool, err := keeper.GetPool(ctx, coin.Asset)
				if err != nil {
					ctx.Logger().Error("fail to get pool", "error", err)
					continue
				}
				totalValue = totalValue.Add(pool.AssetValueInRune(coin.Amount))
			}
		}

		respVaults[i] = QueryYggdrasilVaults{
			vault, na.Status, na.Bond, totalValue,
		}
	}

	res, err := codec.MarshalJSONIndent(keeper.Cdc(), respVaults)
	if err != nil {
		ctx.Logger().Error("fail to marshal vaults response to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal response to json")
	}

	return res, nil
}

func queryVaultsPubkeys(ctx sdk.Context, keeper Keeper) ([]byte, sdk.Error) {
	var resp struct {
		Asgard    common.PubKeys `json:"asgard"`
		Yggdrasil common.PubKeys `json:"yggdrasil"`
	}
	resp.Asgard = make(common.PubKeys, 0)
	resp.Yggdrasil = make(common.PubKeys, 0)
	iter := keeper.GetVaultIterator(ctx)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var vault Vault
		if err := keeper.Cdc().UnmarshalBinaryBare(iter.Value(), &vault); err != nil {
			ctx.Logger().Error("fail to unmarshal yggdrasil", "error", err)
			return nil, sdk.ErrInternal("fail to unmarshal yggdrasil")
		}
		if vault.Status == ActiveVault {
			if vault.IsYggdrasil() {
				resp.Yggdrasil = append(resp.Yggdrasil, vault.PubKey)
			} else if vault.IsAsgard() {
				resp.Asgard = append(resp.Asgard, vault.PubKey)
			}
		}
	}
	res, err := codec.MarshalJSONIndent(keeper.Cdc(), resp)
	if err != nil {
		ctx.Logger().Error("fail to marshal pubkeys response to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal response to json")
	}
	return res, nil
}

func queryVaultData(ctx sdk.Context, keeper Keeper) ([]byte, sdk.Error) {
	data, err := keeper.GetVaultData(ctx)
	if err != nil {
		ctx.Logger().Error("fail to get vault", "error", err)
		return nil, sdk.ErrInternal("fail to get vault")
	}

	if common.RuneAsset().Chain.Equals(common.THORChain) {
		data.TotalReserve = keeper.GetRuneBalaceOfModule(ctx, ReserveName)
	}

	res, err := codec.MarshalJSONIndent(keeper.Cdc(), data)
	if err != nil {
		ctx.Logger().Error("fail to marshal vault data to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal response to json")
	}
	return res, nil
}

func queryPoolAddresses(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	active, err := keeper.GetAsgardVaultsByStatus(ctx, ActiveVault)
	if err != nil {
		ctx.Logger().Error("fail to get active vaults", "error", err)
		return nil, sdk.ErrInternal("fail to get active vaults")
	}

	type address struct {
		Chain   common.Chain   `json:"chain"`
		PubKey  common.PubKey  `json:"pub_key"`
		Address common.Address `json:"address"`
	}

	var resp struct {
		Current []address `json:"current"`
	}

	if len(active) > 0 {
		// select vault with lowest amount of rune
		vault := active.SelectByMinCoin(common.RuneAsset())
		chains := vault.Chains

		if len(chains) == 0 {
			chains = common.Chains{common.RuneAsset().Chain}
		}

		for _, chain := range chains {
			vaultAddress, err := vault.PubKey.GetAddress(chain)
			if err != nil {
				ctx.Logger().Error("fail to get address for chain", "error", err)
				return nil, sdk.ErrInternal("fail to get address for chain")
			}

			addr := address{
				Chain:   chain,
				PubKey:  vault.PubKey,
				Address: vaultAddress,
			}

			resp.Current = append(resp.Current, addr)
		}
	}

	res, err := codec.MarshalJSONIndent(keeper.Cdc(), resp)
	if err != nil {
		ctx.Logger().Error("fail to marshal current pool address to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal current pool address to json")
	}

	return res, nil
}

func queryNodeAccount(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	nodeAddress := path[0]
	addr, err := sdk.AccAddressFromBech32(nodeAddress)
	if err != nil {
		return nil, sdk.ErrUnknownRequest("invalid account address")
	}

	nodeAcc, err := keeper.GetNodeAccount(ctx, addr)
	if err != nil {
		return nil, sdk.ErrInternal("fail to get node accounts")
	}

	slashPts, err := keeper.GetNodeAccountSlashPoints(ctx, addr)
	if err != nil {
		return nil, sdk.ErrInternal("fail to get node slash points")
	}

	result := NewQueryNodeAccount(nodeAcc)
	result.SlashPoints = slashPts
	res, err := codec.MarshalJSONIndent(keeper.Cdc(), result)
	if err != nil {
		ctx.Logger().Error("fail to marshal node account to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal node account to json")
	}

	return res, nil
}

func queryNodeAccounts(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	nodeAccounts, err := keeper.ListNodeAccountsWithBond(ctx)
	if err != nil {
		return nil, sdk.ErrInternal("fail to get node accounts")
	}

	result := make([]QueryNodeAccount, len(nodeAccounts))
	for i, na := range nodeAccounts {
		slashPts, err := keeper.GetNodeAccountSlashPoints(ctx, na.NodeAddress)
		if err != nil {
			return nil, sdk.ErrInternal("fail to get node slash points")
		}
		result[i] = NewQueryNodeAccount(na)
		result[i].SlashPoints = slashPts
	}

	res, err := codec.MarshalJSONIndent(keeper.Cdc(), result)
	if err != nil {
		ctx.Logger().Error("fail to marshal observers to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal observers to json")
	}

	return res, nil
}

// queryObservers will only return all the active accounts
func queryObservers(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	activeAccounts, err := keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		return nil, sdk.ErrInternal("fail to get node account iterator")
	}
	result := make([]string, 0, len(activeAccounts))
	for _, item := range activeAccounts {
		result = append(result, item.NodeAddress.String())
	}
	res, err := codec.MarshalJSONIndent(keeper.Cdc(), result)
	if err != nil {
		ctx.Logger().Error("fail to marshal observers to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal observers to json")
	}

	return res, nil
}

func queryObserver(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	observerAddr := path[0]
	addr, err := sdk.AccAddressFromBech32(observerAddr)
	if err != nil {
		return nil, sdk.ErrUnknownRequest("invalid account address")
	}

	nodeAcc, err := keeper.GetNodeAccount(ctx, addr)
	if err != nil {
		return nil, sdk.ErrInternal("fail to get node account")
	}
	res, err := codec.MarshalJSONIndent(keeper.Cdc(), nodeAcc)
	if err != nil {
		ctx.Logger().Error("fail to marshal node account to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal node account to json")
	}

	return res, nil
}

// queryStakers
func queryStakers(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	asset, err := common.NewAsset(path[0])
	if err != nil {
		ctx.Logger().Error("fail to get parse asset", "error", err)
		return nil, sdk.ErrInternal("fail to parse asset")
	}
	var stakers []Staker
	iterator := keeper.GetStakerIterator(ctx, asset)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var staker Staker
		keeper.Cdc().MustUnmarshalBinaryBare(iterator.Value(), &staker)
		stakers = append(stakers, staker)
	}
	res, err := codec.MarshalJSONIndent(keeper.Cdc(), stakers)
	if err != nil {
		ctx.Logger().Error("fail to marshal stakers to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal stakers to json")
	}
	return res, nil
}

// nolint: unparam
func queryPool(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	asset, err := common.NewAsset(path[0])
	if err != nil {
		ctx.Logger().Error("fail to parse asset", "error", err)
		return nil, sdk.ErrInternal("Could not parse asset")
	}

	active, err := keeper.GetAsgardVaultsByStatus(ctx, ActiveVault)
	if err != nil {
		ctx.Logger().Error("fail to get active vaults", "error", err)
		return nil, sdk.ErrInternal("fail to get active vaults")
	}

	vault := active.SelectByMinCoin(asset)
	if vault.IsEmpty() {
		return nil, sdk.ErrInternal("Could not find active asgard vault")
	}

	addr, err := vault.PubKey.GetAddress(asset.Chain)
	if err != nil {
		return nil, sdk.ErrInternal("fail to get chain pool address")
	}

	pool, err := keeper.GetPool(ctx, asset)
	if err != nil {
		ctx.Logger().Error("fail to get pool", "error", err)
		return nil, sdk.ErrInternal("Could not get pool")
	}
	if pool.Empty() {
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("pool: %s doesn't exist", path[0]))
	}

	pool.PoolAddress = addr
	res, err := codec.MarshalJSONIndent(keeper.Cdc(), pool)
	if err != nil {
		return nil, sdk.ErrInternal("could not marshal result to JSON")
	}
	return res, nil
}

func queryPools(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	pools := QueryResPools{}
	iterator := keeper.GetPoolIterator(ctx)

	active, err := keeper.GetAsgardVaultsByStatus(ctx, ActiveVault)
	if err != nil {
		ctx.Logger().Error("fail to get active vaults", "error", err)
		return nil, sdk.ErrInternal("fail to get active vaults")
	}

	for ; iterator.Valid(); iterator.Next() {
		var pool Pool
		if err := keeper.Cdc().UnmarshalBinaryBare(iterator.Value(), &pool); err != nil {
			return nil, sdk.ErrInternal("Unmarshl: Pool")
		}

		// ignore pool if no stake units
		if pool.PoolUnits.IsZero() {
			continue
		}

		vault := active.SelectByMinCoin(pool.Asset)
		if vault.IsEmpty() {
			return nil, sdk.ErrInternal("Could not find active asgard vault")
		}
		addr, err := vault.PubKey.GetAddress(pool.Asset.Chain)
		if err != nil {
			return nil, sdk.ErrInternal("Could get address of chain")
		}

		pool.PoolAddress = addr
		pools = append(pools, pool)
	}
	res, err := codec.MarshalJSONIndent(keeper.Cdc(), pools)
	if err != nil {
		return nil, sdk.ErrInternal("could not marshal pools result to json")
	}
	return res, nil
}

func queryTxIn(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	hash, err := common.NewTxID(path[0])
	if err != nil {
		ctx.Logger().Error("fail to parse tx id", "error", err)
		return nil, sdk.ErrInternal("fail to parse tx id")
	}
	voter, err := keeper.GetObservedTxVoter(ctx, hash)
	if err != nil {
		ctx.Logger().Error("fail to get observed tx voter", "error", err)
		return nil, sdk.ErrInternal("fail to get observed tx voter")
	}

	nodeAccounts, err := keeper.ListActiveNodeAccounts(ctx)
	if err != nil {
		return nil, sdk.ErrInternal("fail to get node accounts")
	}
	res, err := codec.MarshalJSONIndent(keeper.Cdc(), voter.GetTx(nodeAccounts))
	if err != nil {
		ctx.Logger().Error("fail to marshal tx hash to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal tx hash to json")
	}
	return res, nil
}

func queryKeygen(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var err error
	height, err := strconv.ParseInt(path[0], 0, 64)
	if err != nil {
		ctx.Logger().Error("fail to parse block height", "error", err)
		return nil, sdk.ErrInternal("fail to parse block height")
	}

	if height > ctx.BlockHeight() {
		return nil, sdk.ErrInternal("block height not available yet")
	}

	keygenBlock, err := keeper.GetKeygenBlock(ctx, height)
	if err != nil {
		ctx.Logger().Error("fail to get keygen block", "error", err)
		return nil, sdk.ErrInternal("fail to get keygen block")
	}

	if len(path) > 1 {
		pk, err := common.NewPubKey(path[1])
		if err != nil {
			ctx.Logger().Error("fail to parse pubkey", "error", err)
			return nil, sdk.ErrInternal("fail to parse pubkey")
		}
		// only return those keygen contains the request pub key
		newKeygenBlock := NewKeygenBlock(keygenBlock.Height)
		for _, keygen := range keygenBlock.Keygens {
			if keygen.Members.Contains(pk) {
				newKeygenBlock.Keygens = append(newKeygenBlock.Keygens, keygen)
			}
		}
		keygenBlock = newKeygenBlock
	}

	res, err := codec.MarshalJSONIndent(keeper.Cdc(), keygenBlock)
	if err != nil {
		ctx.Logger().Error("fail to marshal keygen block to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal keygen block to json")
	}
	return res, nil
}

func queryKeysign(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var err error
	height, err := strconv.ParseInt(path[0], 0, 64)
	if err != nil {
		ctx.Logger().Error("fail to parse block height", "error", err)
		return nil, sdk.ErrInternal("fail to parse block height")
	}

	if height > ctx.BlockHeight() {
		return nil, sdk.ErrInternal("block height not available yet")
	}

	pk := common.EmptyPubKey
	if len(path) > 1 {
		pk, err = common.NewPubKey(path[1])
		if err != nil {
			ctx.Logger().Error("fail to parse pubkey", "error", err)
			return nil, sdk.ErrInternal("fail to parse pubkey")
		}
	}
	txs, err := keeper.GetTxOut(ctx, height)
	if err != nil {
		ctx.Logger().Error("fail to get tx out array from key value store", "error", err)
		return nil, sdk.ErrInternal("fail to get tx out array from key value store")
	}

	if !pk.IsEmpty() {
		newTxs := &TxOut{
			Height: txs.Height,
		}
		for _, tx := range txs.TxArray {
			if pk.Equals(tx.VaultPubKey) {
				newTxs.TxArray = append(newTxs.TxArray, tx)
			}
		}
		txs = newTxs
	}

	out := make(map[common.Chain]ResTxOut, 0)
	for _, item := range txs.TxArray {
		res, ok := out[item.Chain]
		if !ok {
			res = ResTxOut{
				Height:  txs.Height,
				Chain:   item.Coin.Asset.Chain,
				TxArray: make([]TxOutItem, 0),
			}
		}
		res.TxArray = append(res.TxArray, *item)
		out[item.Chain] = res
	}

	res, err := codec.MarshalJSONIndent(keeper.Cdc(), QueryResTxOut{
		Chains: out,
	})
	if err != nil {
		ctx.Logger().Error("fail to marshal tx hash to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal tx hash to json")
	}
	return res, nil
}

func isIncludeAllEvents(u *url.URL) bool {
	if u == nil {
		return false
	}
	values, ok := u.Query()["include"]
	if !ok {
		return false
	}
	for _, value := range values {
		if value == "all" {
			return true
		}
	}
	return false
}

func getEventStatusFromQuery(u *url.URL) EventStatuses {
	var result EventStatuses
	if u == nil {
		return result
	}
	values, ok := u.Query()["include"]
	if !ok {
		return result
	}
	return GetEventStatuses(values)
}

func queryEventsByTxHash(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	txID, err := common.NewTxID(path[0])
	if err != nil {
		ctx.Logger().Error("fail to discover tx hash", "error", err)
		return nil, sdk.ErrInternal("fail to discover tx hash")
	}
	eventIDs, err := keeper.GetEventsIDByTxHash(ctx, txID)
	if err != nil {
		errMsg := fmt.Sprintf("fail to get event ids by txhash(%s)", txID.String())
		ctx.Logger().Error(errMsg)
		return nil, sdk.ErrInternal(errMsg)
	}
	limit := 100 // limit the number of events, aka pagination
	if len(eventIDs) > 100 {
		eventIDs = eventIDs[len(eventIDs)-limit:]
	}
	events := make(Events, 0, len(eventIDs))
	for _, id := range eventIDs {
		event, err := keeper.GetEvent(ctx, id)
		if err != nil {
			errMsg := fmt.Sprintf("fail to get event(%d)", id)
			return nil, sdk.ErrInternal(errMsg)
		}

		if event.Empty() {
			break
		}
		events = append(events, event)
	}

	res, err := codec.MarshalJSONIndent(keeper.Cdc(), events)
	if err != nil {
		ctx.Logger().Error("fail to marshal events to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal events to json")
	}
	return res, nil
}

func queryCompEvents(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	id, err := strconv.ParseInt(path[0], 10, 64)
	if err != nil {
		ctx.Logger().Error("fail to discover id number", "error", err)
		return nil, sdk.ErrInternal("fail to discover id number")
	}

	chain := common.EmptyChain
	if len(path) > 1 && len(path[1]) > 0 {
		chain, err = common.NewChain(path[1])
		if err != nil {
			ctx.Logger().Error("fail to discover chain name", "error", err)
			return nil, sdk.ErrInternal("fail to discover chain name")
		}
	}

	u, err := getURLFromData(req.Data)
	if err != nil {
		ctx.Logger().Error(err.Error())
	}
	all := isIncludeAllEvents(u)
	es := getEventStatusFromQuery(u)
	limit := int64(100) // limit the number of events, aka pagination
	events := make(Events, 0)
	for i := id; i <= id+limit; i++ {
		event, _ := keeper.GetEvent(ctx, i)
		if all {
			events = append(events, event)
			continue
		}

		// discover the chain event
		// if no out txs, use intx chain
		// if out txs, use first non-rune chain, else use rune chain
		evtChain := event.InTx.Chain
		if len(event.OutTxs) > 0 {
			evtChain = common.RuneAsset().Chain
			for _, outTx := range event.OutTxs {
				if !outTx.Chain.IsEmpty() && !evtChain.Equals(outTx.Chain) {
					evtChain = outTx.Chain
					break
				}
			}
		}
		// if event is pending, get the chain event from memo
		if event.Status == EventPending {
			memo, _ := ParseMemo(event.InTx.Memo)
			asset := memo.GetAsset()
			if asset.Chain != "" {
				evtChain = asset.Chain
			}
		}

		if !chain.IsEmpty() && !evtChain.Equals(chain) && !evtChain.IsEmpty() {
			continue
		}
		if evtChain.IsEmpty() && !chain.IsBNB() && !chain.IsEmpty() {
			continue
		}
		if event.Empty() {
			continue
		}
		if len(es) == 0 {
			if event.Status == EventPending {
				break
			}
			events = append(events, event)
		} else {
			if es.Contains(event.Status) {
				events = append(events, event)
			}
		}

	}

	res, err := codec.MarshalJSONIndent(keeper.Cdc(), events)
	if err != nil {
		ctx.Logger().Error("fail to marshal events to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal events to json")
	}
	return res, nil
}

func queryHeights(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	chain := common.BNBChain
	if len(path[0]) > 0 {
		var err error
		chain, err = common.NewChain(path[0])
		if err != nil {
			ctx.Logger().Error("fail to retrieve chain", "error", err)
			return nil, sdk.ErrInternal("fail to retrieve chain")
		}
	}
	chainHeight, err := keeper.GetLastChainHeight(ctx, chain)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}

	signed, err := keeper.GetLastSignedHeight(ctx)
	if err != nil {
		return nil, sdk.ErrInternal(err.Error())
	}

	res, err := codec.MarshalJSONIndent(keeper.Cdc(), QueryResHeights{
		Chain:            chain,
		LastChainHeight:  chainHeight,
		LastSignedHeight: signed,
		Statechain:       ctx.BlockHeight(),
	})
	if err != nil {
		ctx.Logger().Error("fail to marshal events to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal events to json")
	}
	return res, nil
}

// queryTSSSigner
func queryTSSSigners(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	vaultPubKey := path[0]
	if len(vaultPubKey) == 0 {
		ctx.Logger().Error("empty vault pub key")
		return nil, sdk.ErrUnknownRequest("empty pool pub key")
	}
	pk, err := common.NewPubKey(vaultPubKey)
	if err != nil {
		ctx.Logger().Error("fail to parse pool pub key", "error", err)
		return nil, sdk.ErrUnknownRequest("invalid pool pub key")
	}

	// seed is the current block height, rounded down to the nearest 100th
	// This helps keep the selected nodes to be the same across blocks, but
	// also change immediately if we have a change in which nodes are active
	seed := ctx.BlockHeight() / 100

	accountAddrs, err := keeper.GetObservingAddresses(ctx)
	if err != nil {
		ctx.Logger().Error("fail to get observing addresses", "error", err)
		return nil, sdk.ErrInternal("fail to get observing addresses")
	}

	vault, err := keeper.GetVault(ctx, pk)
	if err != nil {
		ctx.Logger().Error("fail to get vault", "error", err)
		return nil, sdk.ErrInternal("fail to get vault")
	}
	signers := vault.Membership
	threshold, err := GetThreshold(len(vault.Membership))
	if err != nil {
		ctx.Logger().Error("fail to get threshold", "error", err)
		return nil, sdk.ErrInternal("fail to get threshold")
	}
	totalObservingAccounts := len(accountAddrs)
	if totalObservingAccounts > 0 && totalObservingAccounts >= threshold {
		signers, err = vault.GetMembers(accountAddrs)
		if err != nil {
			ctx.Logger().Error("fail to get signers", "error", err)
			return nil, sdk.ErrInternal("fail to get signers")
		}
	}
	// if we don't have enough signer
	if len(signers) < threshold {
		signers = vault.Membership
	}
	// if there are 9 nodes in total , it need 6 nodes to sign a message
	// 3 signer send request to thorchain at block height 100
	// another 3 signer send request to thorchain at block height 101
	// in this case we get into trouble ,they get different results, key sign is going to fail
	signerParty, err := ChooseSignerParty(signers, seed, len(vault.Membership))
	if err != nil {
		ctx.Logger().Error("fail to choose signer party members", "error", err)
		return nil, sdk.ErrInternal("fail to choose signer party members")
	}
	res, err := codec.MarshalJSONIndent(keeper.Cdc(), signerParty)
	if err != nil {
		ctx.Logger().Error("fail to marshal to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal to json")
	}

	return res, nil
}

func queryConstantValues(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	ver := keeper.GetLowestActiveVersion(ctx)
	constAccessor := constants.GetConstantValues(ver)
	res, err := codec.MarshalJSONIndent(keeper.Cdc(), constAccessor)
	if err != nil {
		ctx.Logger().Error("fail to marshal constant values to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal constant values to json")
	}
	return res, nil
}

func queryMimirValues(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	values := make(map[string]int64, 0)
	iter := keeper.GetMimirIterator(ctx)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var value int64
		if err := keeper.Cdc().UnmarshalBinaryBare(iter.Value(), &value); err != nil {
			ctx.Logger().Error("fail to unmarshal mimir attribute", "error", err)
			return nil, sdk.ErrInternal("fail to unmarshal mimir attribute")
		}
		values[string(iter.Key())] = value
	}
	res, err := codec.MarshalJSONIndent(keeper.Cdc(), values)
	if err != nil {
		ctx.Logger().Error("fail to marshal mimir values to json", "error", err)
		return nil, sdk.ErrInternal("fail to marshal mimir values to json")
	}
	return res, nil
}

func queryBan(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	addr, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		ctx.Logger().Error("invalid node address", "error", err)
		return nil, sdk.ErrInternal("invalid node address")
	}

	ban, err := keeper.GetBanVoter(ctx, addr)
	if err != nil {
		ctx.Logger().Error("fail to get ban voter", "error", err)
		return nil, sdk.ErrInternal("fail to get ban voter")
	}

	res, err := codec.MarshalJSONIndent(keeper.Cdc(), ban)
	if err != nil {
		ctx.Logger().Error("fail to marshal ban voter to json", "error", err)
		return nil, sdk.ErrInternal("fail to ban voter to json")
	}
	return res, nil
}
