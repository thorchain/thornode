package thorchain

import (
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
	validatorMgr := NewVersionedValidatorMgr(keeper, versionedTxOutStoreDummy, versionedVaultMgrDummy)

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
	validatorMgr := NewVersionedValidatorMgr(keeper, versionedTxOutStoreDummy, versionedVaultMgrDummy)

	querier := NewQuerier(keeper, validatorMgr)
	path := []string{"pools"}

	pubKey := GetRandomPubKey()
	asgard := NewVault(ctx.BlockHeight(), ActiveVault, AsgardVault, pubKey, common.Chains{common.BNBChain})
	c.Assert(keeper.SetVault(ctx, asgard), IsNil)

	poolBNB := Pool{
		Asset:     common.BNBAsset,
		PoolUnits: sdk.NewUint(100),
	}
	poolBTC := Pool{
		Asset:     common.BTCAsset,
		PoolUnits: sdk.NewUint(0),
	}
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
	validatorMgr := NewVersionedValidatorMgr(keeper, versionedTxOutStoreDummy, versionedVaultMgrDummy)

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
