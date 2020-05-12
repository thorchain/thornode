package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
)

type HandlerObservedTxOutSuite struct{}

type TestObservedTxOutValidateKeeper struct {
	KVStoreDummy
	activeNodeAccount NodeAccount
}

func (k *TestObservedTxOutValidateKeeper) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	if k.activeNodeAccount.NodeAddress.Equals(signer) {
		return k.activeNodeAccount, nil
	}
	return NodeAccount{}, nil
}

var _ = Suite(&HandlerObservedTxOutSuite{})

func (s *HandlerObservedTxOutSuite) TestValidate(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)
	activeNodeAccount := GetRandomNodeAccount(NodeActive)

	keeper := &TestObservedTxOutValidateKeeper{
		activeNodeAccount: activeNodeAccount,
	}

	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(w.versionedTxOutStore)
	versionedGasMgr := NewVersionedGasMgr()
	versionedObMgr := NewDummyVersionedObserverMgr()
	versionedEventManagerDummy := NewDummyVersionedEventMgr()

	handler := NewObservedTxOutHandler(keeper, versionedObMgr, w.versionedTxOutStore, w.validatorMgr, versionedVaultMgrDummy, versionedGasMgr, versionedEventManagerDummy)

	// happy path
	ver := constants.SWVersion
	pk := GetRandomPubKey()
	txs := ObservedTxs{NewObservedTx(GetRandomTx(), 12, pk)}
	txs[0].Tx.FromAddress, err = pk.GetAddress(txs[0].Tx.Coins[0].Asset.Chain)
	c.Assert(err, IsNil)
	msg := NewMsgObservedTxOut(txs, activeNodeAccount.NodeAddress)
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, errInvalidVersion)

	// inactive node account
	msg = NewMsgObservedTxOut(txs, GetRandomBech32Addr())
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, Equals, notAuthorized)

	// invalid msg
	msg = MsgObservedTxOut{}
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

type TestObservedTxOutFailureKeeper struct {
	KVStoreDummy
}

type TestObservedTxOutHandleKeeper struct {
	KVStoreDummy
	nas        NodeAccounts
	na         NodeAccount
	voter      ObservedTxVoter
	yggExists  bool
	ygg        Vault
	height     int64
	pool       Pool
	txOutStore TxOutStore
	observing  []sdk.AccAddress
	gas        []sdk.Uint
}

func (k *TestObservedTxOutHandleKeeper) ListActiveNodeAccounts(_ sdk.Context) (NodeAccounts, error) {
	return k.nas, nil
}

func (k *TestObservedTxOutHandleKeeper) IsActiveObserver(_ sdk.Context, _ sdk.AccAddress) bool {
	return true
}

func (k *TestObservedTxOutHandleKeeper) GetNodeAccountByPubKey(_ sdk.Context, _ common.PubKey) (NodeAccount, error) {
	return k.nas[0], nil
}

func (k *TestObservedTxOutHandleKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	k.na = na
	return nil
}

func (k *TestObservedTxOutHandleKeeper) GetObservedTxVoter(_ sdk.Context, _ common.TxID) (ObservedTxVoter, error) {
	return k.voter, nil
}

func (k *TestObservedTxOutHandleKeeper) SetObservedTxVoter(_ sdk.Context, voter ObservedTxVoter) {
	k.voter = voter
}

func (k *TestObservedTxOutHandleKeeper) VaultExists(_ sdk.Context, _ common.PubKey) bool {
	return k.yggExists
}

func (k *TestObservedTxOutHandleKeeper) GetVault(_ sdk.Context, _ common.PubKey) (Vault, error) {
	return k.ygg, nil
}

func (k *TestObservedTxOutHandleKeeper) SetVault(_ sdk.Context, ygg Vault) error {
	k.ygg = ygg
	return nil
}

func (k *TestObservedTxOutHandleKeeper) GetVaultData(_ sdk.Context) (VaultData, error) {
	return NewVaultData(), nil
}

func (k *TestObservedTxOutHandleKeeper) SetVaultData(_ sdk.Context, _ VaultData) error {
	return nil
}

func (k *TestObservedTxOutHandleKeeper) SetLastChainHeight(_ sdk.Context, _ common.Chain, height int64) error {
	k.height = height
	return nil
}

func (k *TestObservedTxOutHandleKeeper) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return k.pool, nil
}

func (k *TestObservedTxOutHandleKeeper) AddIncompleteEvents(_ sdk.Context, evt Event) error {
	return nil
}

func (k *TestObservedTxOutHandleKeeper) GetTxOut(ctx sdk.Context, _ int64) (*TxOut, error) {
	return k.txOutStore.GetBlockOut(ctx)
}

func (k *TestObservedTxOutHandleKeeper) FindPubKeyOfAddress(_ sdk.Context, _ common.Address, _ common.Chain) (common.PubKey, error) {
	return k.ygg.PubKey, nil
}

func (k *TestObservedTxOutHandleKeeper) SetTxOut(_ sdk.Context, _ *TxOut) error {
	return nil
}

func (k *TestObservedTxOutHandleKeeper) AddObservingAddresses(_ sdk.Context, addrs []sdk.AccAddress) error {
	k.observing = addrs
	return nil
}

func (k *TestObservedTxOutHandleKeeper) UpsertEvent(ctx sdk.Context, event Event) error {
	return nil
}

func (k *TestObservedTxOutHandleKeeper) GetLastEventID(_ sdk.Context) (int64, error) {
	return 0, nil
}

func (k *TestObservedTxOutHandleKeeper) GetIncompleteEvents(_ sdk.Context) (Events, error) {
	return nil, nil
}

func (k *TestObservedTxOutHandleKeeper) SetPool(ctx sdk.Context, pool Pool) error {
	k.pool = pool
	return nil
}

func (k *TestObservedTxOutHandleKeeper) GetGas(ctx sdk.Context, asset common.Asset) ([]sdk.Uint, error) {
	return k.gas, nil
}

func (k *TestObservedTxOutHandleKeeper) SetGas(ctx sdk.Context, asset common.Asset, units []sdk.Uint) {
	k.gas = units
}

func (s *HandlerObservedTxOutSuite) TestHandle(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	ver := constants.SWVersion
	tx := GetRandomTx()
	tx.Memo = fmt.Sprintf("OUTBOUND:%s", tx.ID)
	obTx := NewObservedTx(tx, 12, GetRandomPubKey())
	txs := ObservedTxs{obTx}
	pk := GetRandomPubKey()
	c.Assert(err, IsNil)

	versionedTxOutStoreDummy := NewVersionedTxOutStoreDummy()

	ygg := NewVault(ctx.BlockHeight(), ActiveVault, YggdrasilVault, pk, common.Chains{common.BNBChain})
	ygg.Coins = common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(500)),
		common.NewCoin(common.BNBAsset, sdk.NewUint(200*common.One)),
	}
	keeper := &TestObservedTxOutHandleKeeper{
		nas:   NodeAccounts{GetRandomNodeAccount(NodeActive)},
		voter: NewObservedTxVoter(tx.ID, make(ObservedTxs, 0)),
		pool: Pool{
			Asset:        common.BNBAsset,
			BalanceRune:  sdk.NewUint(200),
			BalanceAsset: sdk.NewUint(300),
		},
		yggExists: true,
		ygg:       ygg,
	}
	txOutStore, err := versionedTxOutStoreDummy.GetTxOutStore(ctx, keeper, ver)
	keeper.txOutStore = txOutStore
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStoreDummy)
	versionedGasMgr := NewVersionedGasMgr()
	versionedObMgr := NewVersionedObserverMgr()
	versionedEventManagerDummy := NewDummyVersionedEventMgr()

	handler := NewObservedTxOutHandler(keeper, versionedObMgr, versionedTxOutStoreDummy, w.validatorMgr, versionedVaultMgrDummy, versionedGasMgr, versionedEventManagerDummy)

	c.Assert(err, IsNil)
	msg := NewMsgObservedTxOut(txs, keeper.nas[0].NodeAddress)
	result := handler.handle(ctx, msg, ver)
	c.Assert(result.IsOK(), Equals, true)
	obMgr, err := versionedObMgr.GetObserverManager(ctx, ver)
	c.Assert(err, IsNil)
	obMgr.EndBlock(ctx, keeper)

	items, err := txOutStore.GetOutboundItems(ctx)
	c.Assert(err, IsNil)
	c.Assert(items, HasLen, 0)
	c.Check(keeper.observing, HasLen, 1)
	// make sure the coin has been subtract from the vault
	c.Check(ygg.Coins.GetCoin(common.BNBAsset).Amount.Equal(sdk.NewUint(19999962499)), Equals, true, Commentf("%d", ygg.Coins.GetCoin(common.BNBAsset).Amount.Uint64()))
}

func (s *HandlerObservedTxOutSuite) TestGasUpdate(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	ver := constants.SWVersion
	tx := GetRandomTx()
	tx.Gas = common.Gas{
		{
			Asset:  common.BNBAsset,
			Amount: sdk.NewUint(475000),
		},
	}
	tx.Memo = fmt.Sprintf("OUTBOUND:%s", tx.ID)
	obTx := NewObservedTx(tx, 12, GetRandomPubKey())
	txs := ObservedTxs{obTx}
	pk := GetRandomPubKey()
	c.Assert(err, IsNil)

	versionedTxOutStoreDummy := NewVersionedTxOutStoreDummy()

	ygg := NewVault(ctx.BlockHeight(), ActiveVault, YggdrasilVault, pk, common.Chains{common.BNBChain})
	ygg.Coins = common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(500)),
		common.NewCoin(common.BNBAsset, sdk.NewUint(200*common.One)),
	}
	keeper := &TestObservedTxOutHandleKeeper{
		nas:   NodeAccounts{GetRandomNodeAccount(NodeActive)},
		voter: NewObservedTxVoter(tx.ID, make(ObservedTxs, 0)),
		pool: Pool{
			Asset:        common.BNBAsset,
			BalanceRune:  sdk.NewUint(200),
			BalanceAsset: sdk.NewUint(300),
		},
		yggExists: true,
		ygg:       ygg,
	}
	txOutStore, err := versionedTxOutStoreDummy.GetTxOutStore(ctx, keeper, ver)
	keeper.txOutStore = txOutStore
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStoreDummy)
	versionedGasMgr := NewVersionedGasMgr()
	versionedObMgr := NewDummyVersionedObserverMgr()
	versionedEventManagerDummy := NewDummyVersionedEventMgr()

	handler := NewObservedTxOutHandler(keeper, versionedObMgr, versionedTxOutStoreDummy, w.validatorMgr, versionedVaultMgrDummy, versionedGasMgr, versionedEventManagerDummy)

	c.Assert(err, IsNil)
	msg := NewMsgObservedTxOut(txs, keeper.nas[0].NodeAddress)
	result := handler.handle(ctx, msg, ver)
	c.Assert(result.IsOK(), Equals, true)
	gas := keeper.gas[0]
	c.Assert(gas.Equal(sdk.NewUint(475000)), Equals, true, Commentf("%+v", gas))
	// revert the gas change , otherwise it messed up the other tests
	gasInfo := common.UpdateGasPrice(common.Tx{}, common.BNBAsset, []sdk.Uint{sdk.NewUint(37500), sdk.NewUint(30000)})
	keeper.SetGas(ctx, common.BNBAsset, gasInfo)
}

func (s *HandlerObservedTxOutSuite) TestHandleStolenFunds(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	ver := constants.SWVersion
	tx := GetRandomTx()
	tx.Memo = "I AM A THIEF!" // bad memo
	obTx := NewObservedTx(tx, 12, GetRandomPubKey())
	obTx.Tx.Coins = common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(300*common.One)),
		common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
	}
	txs := ObservedTxs{obTx}
	pk := GetRandomPubKey()
	c.Assert(err, IsNil)

	na := GetRandomNodeAccount(NodeActive)
	na.Bond = sdk.NewUint(1000000 * common.One)
	na.PubKeySet.Secp256k1 = pk

	versionedTxOutStoreDummy := NewVersionedTxOutStoreDummy()

	ygg := NewVault(ctx.BlockHeight(), ActiveVault, YggdrasilVault, pk, common.Chains{common.BNBChain})
	ygg.Coins = common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(500*common.One)),
		common.NewCoin(common.BNBAsset, sdk.NewUint(200*common.One)),
	}
	keeper := &TestObservedTxOutHandleKeeper{
		nas:   NodeAccounts{na},
		voter: NewObservedTxVoter(tx.ID, make(ObservedTxs, 0)),
		pool: Pool{
			Asset:        common.BNBAsset,
			BalanceRune:  sdk.NewUint(200 * common.One),
			BalanceAsset: sdk.NewUint(300 * common.One),
		},
		yggExists: true,
		ygg:       ygg,
	}
	txOutStore, err := versionedTxOutStoreDummy.GetTxOutStore(ctx, keeper, ver)
	keeper.txOutStore = txOutStore
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStoreDummy)
	versionedGasMgr := NewVersionedGasMgr()
	versionedObMgr := NewDummyVersionedObserverMgr()
	versionedEventManagerDummy := NewDummyVersionedEventMgr()

	handler := NewObservedTxOutHandler(keeper, versionedObMgr, versionedTxOutStoreDummy, w.validatorMgr, versionedVaultMgrDummy, versionedGasMgr, versionedEventManagerDummy)

	c.Assert(err, IsNil)
	msg := NewMsgObservedTxOut(txs, keeper.nas[0].NodeAddress)
	result := handler.handle(ctx, msg, ver)
	c.Assert(result.IsOK(), Equals, true)
	// make sure the coin has been subtract from the vault
	c.Check(ygg.Coins.GetCoin(common.BNBAsset).Amount.Equal(sdk.NewUint(9999962500)), Equals, true, Commentf("%d", ygg.Coins.GetCoin(common.BNBAsset).Amount.Uint64()))
	c.Assert(keeper.na.Bond.LT(sdk.NewUint(1000000*common.One)), Equals, true, Commentf("%d", keeper.na.Bond.Uint64()))
}
