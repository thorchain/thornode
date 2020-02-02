package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type HandlerObservedTxInSuite struct{}

type TestObservedTxInValidateKeeper struct {
	KVStoreDummy
	isActive       bool
	standbyAccount NodeAccount
}

func (k *TestObservedTxInValidateKeeper) GetNodeAccount(_ sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	if addr.Equals(k.standbyAccount.NodeAddress) {
		return k.standbyAccount, nil
	}
	return NodeAccount{}, kaboom
}

func (k *TestObservedTxInValidateKeeper) SetNodeAccount(_ sdk.Context, na NodeAccount) error {
	if na.NodeAddress.Equals(k.standbyAccount.NodeAddress) {
		k.standbyAccount = na
		return nil
	}
	return kaboom
}

func (k *TestObservedTxInValidateKeeper) IsActiveObserver(_ sdk.Context, _ sdk.AccAddress) bool {
	return k.isActive
}

var _ = Suite(&HandlerObservedTxInSuite{})

func (s *HandlerObservedTxInSuite) TestValidate(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)
	standbyAccount := GetRandomNodeAccount(NodeStandby)
	keeper := &TestObservedTxInValidateKeeper{
		isActive:       true,
		standbyAccount: standbyAccount,
	}

	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(w.versionedTxOutStore)
	handler := NewObservedTxInHandler(keeper, w.versionedTxOutStore, w.validatorMgr, versionedVaultMgrDummy)

	// happy path
	ver := semver.MustParse("0.1.0")
	pk := GetRandomPubKey()
	txs := ObservedTxs{NewObservedTx(GetRandomTx(), 12, pk)}
	txs[0].Tx.ToAddress, err = pk.GetAddress(txs[0].Tx.Coins[0].Asset.Chain)
	c.Assert(err, IsNil)
	msg := NewMsgObservedTxIn(txs, GetRandomBech32Addr())
	isNewSigner, err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)
	c.Assert(isNewSigner, Equals, false)

	// invalid version
	isNewSigner, err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, errInvalidVersion)
	c.Assert(isNewSigner, Equals, false)

	// inactive node account
	keeper.isActive = false
	msg = NewMsgObservedTxIn(txs, GetRandomBech32Addr())
	isNewSigner, err = handler.validate(ctx, msg, ver)
	c.Assert(err, Equals, notAuthorized)
	c.Assert(isNewSigner, Equals, false)

	// invalid msg
	msg = MsgObservedTxIn{}
	isNewSigner, err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)
	c.Assert(isNewSigner, Equals, false)

	// test it is signed by a new observer
	msg = NewMsgObservedTxIn(txs, standbyAccount.NodeAddress)
	isNewSigner, err = handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)
	c.Assert(isNewSigner, Equals, true)
	c.Assert(keeper.standbyAccount.ObserverActive, Equals, true)
}

type TestObservedTxInFailureKeeper struct {
	KVStoreDummy
	pool Pool
	evt  Event
}

func (k *TestObservedTxInFailureKeeper) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return k.pool, nil
}

func (k *TestObservedTxInFailureKeeper) UpsertEvent(_ sdk.Context, evt Event) error {
	k.evt = evt
	return nil
}

func (s *HandlerObservedTxInSuite) TestFailure(c *C) {
	ctx, _ := setupKeeperForTest(c)
	// w := getHandlerTestWrapper(c, 1, true, false)

	keeper := &TestObservedTxInFailureKeeper{
		pool: Pool{
			Asset:        common.BNBAsset,
			BalanceRune:  sdk.NewUint(200),
			BalanceAsset: sdk.NewUint(300),
		},
	}
	txOutStore := NewTxStoreDummy()

	tx := NewObservedTx(GetRandomTx(), 12, GetRandomPubKey())
	err := refundTx(ctx, tx, txOutStore, keeper, CodeInvalidMemo, "Invalid memo")
	c.Assert(err, IsNil)
	items, err := txOutStore.GetOutboundItems(ctx)
	c.Assert(err, IsNil)
	c.Check(items, HasLen, 1)
}

type TestObservedTxInHandleKeeper struct {
	KVStoreDummy
	nas       NodeAccounts
	voter     ObservedTxVoter
	yggExists bool
	height    int64
	chains    common.Chains
	pool      Pool
	observing []sdk.AccAddress
	vault     Vault
	txOut     *TxOut
}

func (k *TestObservedTxInHandleKeeper) ListActiveNodeAccounts(_ sdk.Context) (NodeAccounts, error) {
	return k.nas, nil
}

func (k *TestObservedTxInHandleKeeper) GetObservedTxVoter(_ sdk.Context, _ common.TxID) (ObservedTxVoter, error) {
	return k.voter, nil
}

func (k *TestObservedTxInHandleKeeper) SetObservedTxVoter(_ sdk.Context, voter ObservedTxVoter) {
	k.voter = voter
}

func (k *TestObservedTxInHandleKeeper) VaultExists(_ sdk.Context, _ common.PubKey) bool {
	return k.yggExists
}

func (k *TestObservedTxInHandleKeeper) GetChains(_ sdk.Context) (common.Chains, error) {
	return k.chains, nil
}

func (k *TestObservedTxInHandleKeeper) SetChains(_ sdk.Context, chains common.Chains) {
	k.chains = chains
}

func (k *TestObservedTxInHandleKeeper) SetLastChainHeight(_ sdk.Context, _ common.Chain, height int64) error {
	k.height = height
	return nil
}

func (k *TestObservedTxInHandleKeeper) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return k.pool, nil
}

func (k *TestObservedTxInHandleKeeper) AddIncompleteEvents(_ sdk.Context, evt Event) error {
	return nil
}

func (k *TestObservedTxInHandleKeeper) AddObservingAddresses(_ sdk.Context, addrs []sdk.AccAddress) error {
	k.observing = addrs
	return nil
}

func (k *TestObservedTxInHandleKeeper) UpsertEvent(_ sdk.Context, _ Event) error {
	return nil
}

func (k *TestObservedTxInHandleKeeper) GetVault(_ sdk.Context, key common.PubKey) (Vault, error) {
	if k.vault.PubKey.Equals(key) {
		return k.vault, nil
	}
	return GetRandomVault(), kaboom
}

func (k *TestObservedTxInHandleKeeper) SetVault(_ sdk.Context, vault Vault) error {
	if k.vault.PubKey.Equals(vault.PubKey) {
		k.vault = vault
		return nil
	}
	return kaboom
}
func (k *TestObservedTxInHandleKeeper) GetLowestActiveVersion(_ sdk.Context) semver.Version {
	return semver.MustParse("0.1.0")
}
func (k *TestObservedTxInHandleKeeper) IsActiveObserver(_ sdk.Context, addr sdk.AccAddress) bool {
	if addr.Equals(k.nas[0].NodeAddress) {
		return true
	}
	return false
}

func (k *TestObservedTxInHandleKeeper) GetTxOut(ctx sdk.Context, blockHeight int64) (*TxOut, error) {
	if nil != k.txOut && k.txOut.Height == blockHeight {
		return k.txOut, nil
	}
	return nil, kaboom
}
func (k *TestObservedTxInHandleKeeper) SetTxOut(ctx sdk.Context, blockOut *TxOut) error {
	if k.txOut.Height == blockOut.Height {
		k.txOut = blockOut
		return nil
	}
	return kaboom
}

func (s *HandlerObservedTxInSuite) TestHandle(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	ver := semver.MustParse("0.1.0")

	tx := GetRandomTx()
	tx.Memo = "SWAP:BTC.BTC"
	obTx := NewObservedTx(tx, 12, GetRandomPubKey())
	txs := ObservedTxs{obTx}
	pk := GetRandomPubKey()
	txs[0].Tx.ToAddress, err = pk.GetAddress(txs[0].Tx.Coins[0].Asset.Chain)

	vault := GetRandomVault()
	vault.PubKey = obTx.ObservedPubKey

	keeper := &TestObservedTxInHandleKeeper{
		nas:   NodeAccounts{GetRandomNodeAccount(NodeActive)},
		voter: NewObservedTxVoter(tx.ID, make(ObservedTxs, 0)),
		vault: vault,
		pool: Pool{
			Asset:        common.BNBAsset,
			BalanceRune:  sdk.NewUint(200),
			BalanceAsset: sdk.NewUint(300),
		},
		yggExists: true,
	}
	versionedTxOutStore := NewVersionedTxOutStoreDummy()
	txOutStore, err := versionedTxOutStore.GetTxOutStore(keeper, ver)
	c.Assert(err, IsNil)
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStore)
	handler := NewObservedTxInHandler(keeper, versionedTxOutStore, w.validatorMgr, versionedVaultMgrDummy)

	c.Assert(err, IsNil)
	msg := NewMsgObservedTxIn(txs, keeper.nas[0].NodeAddress)
	result := handler.handle(ctx, msg, ver)
	c.Assert(result.IsOK(), Equals, true)
	items, err := txOutStore.GetOutboundItems(ctx)
	c.Assert(err, IsNil)
	c.Check(items, HasLen, 1)
	c.Check(keeper.observing, HasLen, 1)
	c.Check(keeper.height, Equals, int64(12))
	c.Check(keeper.chains, HasLen, 1)
	c.Check(keeper.chains[0].Equals(common.BNBChain), Equals, true)
	bnbCoin := keeper.vault.Coins.GetCoin(common.BNBAsset)
	c.Assert(bnbCoin.Amount.Equal(sdk.OneUint()), Equals, true)
}

// Test migrate memo
func (s *HandlerObservedTxInSuite) TestMigrateMemo(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)
	ver := semver.MustParse("0.1.0")

	vault := GetRandomVault()
	addr, err := vault.PubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	newVault := GetRandomVault()
	txout := NewTxOut(12)
	newVaultAddr, err := newVault.PubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)

	txout.TxArray = append(txout.TxArray, &TxOutItem{
		Chain:       common.BNBChain,
		InHash:      common.BlankTxID,
		ToAddress:   newVaultAddr,
		VaultPubKey: vault.PubKey,
		Coin:        common.NewCoin(common.BNBAsset, sdk.NewUint(1024)),
		Memo:        NewMigrateMemo(1).String(),
	})
	tx := NewObservedTx(common.Tx{
		ID:    GetRandomTxHash(),
		Chain: common.BNBChain,
		Coins: common.Coins{
			common.NewCoin(common.BNBAsset, sdk.NewUint(1024))},
		Memo:        NewMigrateMemo(12).String(),
		FromAddress: addr,
		ToAddress:   newVaultAddr,
		Gas:         common.BNBGasFeeSingleton,
	}, 13, vault.PubKey)

	txs := ObservedTxs{tx}
	keeper := &TestObservedTxInHandleKeeper{
		nas:   NodeAccounts{GetRandomNodeAccount(NodeActive)},
		voter: NewObservedTxVoter(tx.Tx.ID, make(ObservedTxs, 0)),
		vault: vault,
		pool: Pool{
			Asset:        common.BNBAsset,
			BalanceRune:  sdk.NewUint(200),
			BalanceAsset: sdk.NewUint(300),
		},
		yggExists: true,
		txOut:     txout,
	}
	versionedTxOutStore := NewVersionedTxOutStoreDummy()
	c.Assert(err, IsNil)
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStore)
	handler := NewObservedTxInHandler(keeper, versionedTxOutStore, w.validatorMgr, versionedVaultMgrDummy)

	c.Assert(err, IsNil)
	msg := NewMsgObservedTxIn(txs, keeper.nas[0].NodeAddress)
	result := handler.handle(ctx, msg, ver)
	c.Assert(result.IsOK(), Equals, true)
}
