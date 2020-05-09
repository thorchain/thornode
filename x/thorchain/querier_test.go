package thorchain

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

type QuerierSuite struct{}

var _ = Suite(&QuerierSuite{})

type TestQuerierKeeper struct {
	KVStoreDummy
	txOut *TxOut
}

func (k *TestQuerierKeeper) GetTxOut(_ sdk.Context, _ int64) (*TxOut, error) {
	return k.txOut, nil
}

func (s *QuerierSuite) TestQueryKeysign(c *C) {
	ctx, _ := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(12)

	pk := GetRandomPubKey()
	toAddr := GetRandomBNBAddress()
	txOut := NewTxOut(1)
	txOutItem := &TxOutItem{
		Chain:       common.BNBChain,
		VaultPubKey: pk,
		ToAddress:   toAddr,
		InHash:      GetRandomTxHash(),
		Coin:        common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
	}
	txOut.TxArray = append(txOut.TxArray, txOutItem)
	keeper := &TestQuerierKeeper{
		txOut: txOut,
	}

	versionedTxOutStoreDummy := NewVersionedTxOutStoreDummy()
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStoreDummy)
	versionedEventManagerDummy := NewDummyVersionedEventMgr()

	validatorMgr := NewVersionedValidatorMgr(keeper, versionedTxOutStoreDummy, versionedVaultMgrDummy, versionedEventManagerDummy)

	querier := NewQuerier(keeper, validatorMgr)

	path := []string{
		"keysign",
		"5",
		pk.String(),
	}
	res, err := querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)
	c.Assert(res, NotNil)
}

func (s *QuerierSuite) TestQueryPool(c *C) {
	ctx, keeper := setupKeeperForTest(c)

	versionedTxOutStoreDummy := NewVersionedTxOutStoreDummy()
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStoreDummy)
	versionedEventManagerDummy := NewDummyVersionedEventMgr()

	validatorMgr := NewVersionedValidatorMgr(keeper, versionedTxOutStoreDummy, versionedVaultMgrDummy, versionedEventManagerDummy)

	querier := NewQuerier(keeper, validatorMgr)
	path := []string{"pools"}

	pubKey := GetRandomPubKey()
	asgard := NewVault(ctx.BlockHeight(), ActiveVault, AsgardVault, pubKey, common.Chains{common.BNBChain})
	c.Assert(keeper.SetVault(ctx, asgard), IsNil)

	poolBNB := NewPool()
	poolBNB.Asset = common.BNBAsset
	poolBNB.PoolUnits = sdk.NewUint(100)

	poolBTC := NewPool()
	poolBTC.Asset = common.BTCAsset
	poolBTC.PoolUnits = sdk.NewUint(0)

	err := keeper.SetPool(ctx, poolBNB)
	c.Assert(err, IsNil)

	err = keeper.SetPool(ctx, poolBTC)
	c.Assert(err, IsNil)

	res, err := querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)

	var out types.QueryResPools
	err = keeper.Cdc().UnmarshalJSON(res, &out)
	c.Assert(err, IsNil)
	c.Assert(len(out), Equals, 1)

	poolBTC.PoolUnits = sdk.NewUint(100)
	err = keeper.SetPool(ctx, poolBTC)
	c.Assert(err, IsNil)

	res, err = querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)

	err = keeper.Cdc().UnmarshalJSON(res, &out)
	c.Assert(err, IsNil)
	c.Assert(len(out), Equals, 2)
}

func (s *QuerierSuite) TestQueryNodeAccounts(c *C) {
	ctx, keeper := setupKeeperForTest(c)

	versionedTxOutStoreDummy := NewVersionedTxOutStoreDummy()
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStoreDummy)
	versionedEventManagerDummy := NewDummyVersionedEventMgr()

	validatorMgr := NewVersionedValidatorMgr(keeper, versionedTxOutStoreDummy, versionedVaultMgrDummy, versionedEventManagerDummy)

	querier := NewQuerier(keeper, validatorMgr)
	path := []string{"nodeaccounts"}

	signer := GetRandomBech32Addr()
	bondAddr := GetRandomBNBAddress()
	emptyPubKeySet := common.PubKeySet{}
	bond := sdk.NewUint(common.One * 100)
	nodeAccount := NewNodeAccount(signer, NodeActive, emptyPubKeySet, "", bond, bondAddr, ctx.BlockHeight())
	c.Assert(keeper.SetNodeAccount(ctx, nodeAccount), IsNil)

	res, err := querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)

	var out types.NodeAccounts
	err1 := keeper.Cdc().UnmarshalJSON(res, &out)
	c.Assert(err1, IsNil)
	c.Assert(len(out), Equals, 1)

	signer = GetRandomBech32Addr()
	bondAddr = GetRandomBNBAddress()
	emptyPubKeySet = common.PubKeySet{}
	bond = sdk.NewUint(common.One * 200)
	nodeAccount2 := NewNodeAccount(signer, NodeActive, emptyPubKeySet, "", bond, bondAddr, ctx.BlockHeight())
	c.Assert(keeper.SetNodeAccount(ctx, nodeAccount2), IsNil)

	res, err = querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)

	err1 = keeper.Cdc().UnmarshalJSON(res, &out)
	c.Assert(err1, IsNil)
	c.Assert(len(out), Equals, 2)

	nodeAccount2.Bond = sdk.NewUint(0)
	c.Assert(keeper.SetNodeAccount(ctx, nodeAccount2), IsNil)

	res, err = querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)

	err1 = keeper.Cdc().UnmarshalJSON(res, &out)
	c.Assert(err1, IsNil)
	c.Assert(len(out), Equals, 1)
}

func (s *QuerierSuite) TestQueryCompEvents(c *C) {
	ctx, keeper := setupKeeperForTest(c)

	versionedTxOutStoreDummy := NewVersionedTxOutStoreDummy()
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStoreDummy)
	versionedEventManagerDummy := NewDummyVersionedEventMgr()

	validatorMgr := NewVersionedValidatorMgr(keeper, versionedTxOutStoreDummy, versionedVaultMgrDummy, versionedEventManagerDummy)

	querier := NewQuerier(keeper, validatorMgr)
	path := []string{"comp_events_chain", "1", "BNB"}

	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	txIn := common.NewTx(
		txID,
		GetRandomBNBAddress(),
		GetRandomBNBAddress(),
		common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(320000000)),
			common.NewCoin(common.RuneAsset(), sdk.NewUint(420000000)),
		},
		BNBGasFeeSingleton,
		"SWAP:BNB.BNB",
	)
	stake := NewEventStake(
		common.BNBAsset,
		sdk.NewUint(5),
		txIn,
	)
	stakeBytes, _ := json.Marshal(stake)
	evt := NewEvent(
		stake.Type(),
		12,
		txIn,
		stakeBytes,
		EventSuccess,
	)
	c.Assert(keeper.UpsertEvent(ctx, evt), IsNil)

	res, err := querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)

	var out Events
	err = keeper.Cdc().UnmarshalJSON(res, &out)
	c.Assert(err, IsNil)
	c.Assert(len(out), Equals, 1)

	// add empty tx in out tx and should be returned still
	// because the in tx chain match
	evt.OutTxs = common.Txs{common.Tx{ID: common.BlankTxID}}
	c.Assert(keeper.UpsertEvent(ctx, evt), IsNil)

	res, err = querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)

	err = keeper.Cdc().UnmarshalJSON(res, &out)
	c.Assert(err, IsNil)
	c.Assert(len(out), Equals, 2)

	// add new event with BTC chain in out txs
	evt.OutTxs = common.Txs{common.Tx{Chain: common.BTCChain, ID: common.BlankTxID}}
	c.Assert(keeper.UpsertEvent(ctx, evt), IsNil)

	res, err = querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)

	// BNB events count should be the same and BTC out tx
	// event should be ignored
	err = keeper.Cdc().UnmarshalJSON(res, &out)
	c.Assert(err, IsNil)
	c.Assert(len(out), Equals, 2)

	// query BTC chain event should return only the last event
	path = []string{"comp_events_chain", "1", "BTC"}
	res, err = querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)

	err = keeper.Cdc().UnmarshalJSON(res, &out)
	c.Assert(err, IsNil)
	c.Assert(len(out), Equals, 1)
	c.Assert(out[0].OutTxs[0].Chain.Equals(common.BTCChain), Equals, true)

	// add a gas event with empty chain and make sure its return in BNB chain calls
	evt.InTx.Chain = common.EmptyChain
	evt.OutTxs = nil
	c.Assert(keeper.UpsertEvent(ctx, evt), IsNil)

	// regular complete events without chain should have now all 4 events
	path = []string{"comp_events", "1"}
	res, err = querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)
	err = keeper.Cdc().UnmarshalJSON(res, &out)
	c.Assert(err, IsNil)
	c.Assert(len(out), Equals, 4)

	// BNB events should have 3 events including gas event with empty tx in chain
	path = []string{"comp_events", "1", "BNB"}
	res, err = querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)
	err = keeper.Cdc().UnmarshalJSON(res, &out)
	c.Assert(err, IsNil)
	c.Assert(len(out), Equals, 3)
	c.Assert(out[2].InTx.Chain.IsEmpty(), Equals, true)

	// BTC events should still have the 1 event
	path = []string{"comp_events", "1", "BTC"}
	res, err = querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)
	err = keeper.Cdc().UnmarshalJSON(res, &out)
	c.Assert(err, IsNil)
	c.Assert(len(out), Equals, 1)
	c.Assert(out[0].OutTxs[0].Chain.Equals(common.BTCChain), Equals, true)

	// check regular query complete events still works correctly
	// regular complete events without chain should have now all 4 events
	path = []string{"comp_events", "1"}
	res, err = querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)

	err = keeper.Cdc().UnmarshalJSON(res, &out)
	c.Assert(err, IsNil)
	c.Assert(len(out), Equals, 4)
	c.Assert(out[2].OutTxs[0].Chain.Equals(common.BTCChain), Equals, true)
	c.Assert(out[3].InTx.Chain.IsEmpty(), Equals, true)

	// check call with empty chain id
	path = []string{"comp_events", "1", ""}
	res, err = querier(ctx, path, abci.RequestQuery{})
	c.Assert(err, IsNil)

	err = keeper.Cdc().UnmarshalJSON(res, &out)
	c.Assert(err, IsNil)
	c.Assert(len(out), Equals, 4)
	c.Assert(out[2].OutTxs[0].Chain.Equals(common.BTCChain), Equals, true)
	c.Assert(out[3].InTx.Chain.IsEmpty(), Equals, true)
}
