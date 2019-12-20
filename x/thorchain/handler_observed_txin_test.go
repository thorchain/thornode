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

func (k *TestObservedTxInValidateKeeper) GetNodeAccount(ctx sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	if addr.Equals(k.standbyAccount.NodeAddress) {
		return k.standbyAccount, nil
	}
	return NodeAccount{}, kaboom
}
func (k *TestObservedTxInValidateKeeper) SetNodeAccount(ctx sdk.Context, na NodeAccount) error {
	if na.NodeAddress.Equals(k.standbyAccount.NodeAddress) {
		k.standbyAccount = na
		return nil
	}
	return kaboom
}

func (k *TestObservedTxInValidateKeeper) IsActiveObserver(ctx sdk.Context, signer sdk.AccAddress) bool {
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

	vaultMgr := NewVaultMgrDummy()
	handler := NewObservedTxInHandler(keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr, vaultMgr)

	// happy path
	ver := semver.MustParse("0.1.0")
	pk := GetRandomPubKey()
	txs := ObservedTxs{NewObservedTx(GetRandomTx(), sdk.NewUint(12), pk)}
	txs[0].Tx.ToAddress, err = pk.GetAddress(txs[0].Tx.Coins[0].Asset.Chain)
	c.Assert(err, IsNil)
	msg := NewMsgObservedTxIn(txs, GetRandomBech32Addr())
	isNewSigner, err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)
	c.Assert(isNewSigner, Equals, false)

	// invalid version
	isNewSigner, err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)
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

func (k *TestObservedTxInFailureKeeper) AddIncompleteEvents(_ sdk.Context, evt Event) error {
	k.evt = evt
	return nil
}

func (s *HandlerObservedTxInSuite) TestFailure(c *C) {
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	keeper := &TestObservedTxInFailureKeeper{
		pool: Pool{
			Asset:        common.BNBAsset,
			BalanceRune:  sdk.NewUint(200),
			BalanceAsset: sdk.NewUint(300),
		},
	}
	txOutStore := NewTxStoreDummy()

	vaultMgr := NewVaultMgrDummy()
	handler := NewObservedTxInHandler(keeper, txOutStore, w.poolAddrMgr, w.validatorMgr, vaultMgr)
	tx := NewObservedTx(GetRandomTx(), sdk.NewUint(12), GetRandomPubKey())

	err := handler.inboundFailure(ctx, tx)
	c.Assert(err, IsNil)
	c.Check(txOutStore.GetOutboundItems(), HasLen, 1)
	c.Check(keeper.evt.Empty(), Equals, false, Commentf("%+v", keeper.evt))
}

type TestObservedTxInHandleKeeper struct {
	KVStoreDummy
	nas       NodeAccounts
	voter     ObservedTxVoter
	yggExists bool
	height    sdk.Uint
	chains    common.Chains
	pool      Pool
	observing []sdk.AccAddress
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

func (k *TestObservedTxInHandleKeeper) SetLastChainHeight(_ sdk.Context, _ common.Chain, height sdk.Uint) error {
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

func (s *HandlerObservedTxInSuite) TestHandle(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	ver := semver.MustParse("0.1.0")

	tx := GetRandomTx()
	tx.Memo = "SWAP:BTC.BTC"
	obTx := NewObservedTx(tx, sdk.NewUint(12), GetRandomPubKey())
	txs := ObservedTxs{obTx}
	pk := GetRandomPubKey()
	txs[0].Tx.ToAddress, err = pk.GetAddress(txs[0].Tx.Coins[0].Asset.Chain)

	keeper := &TestObservedTxInHandleKeeper{
		nas:   NodeAccounts{GetRandomNodeAccount(NodeActive)},
		voter: NewObservedTxVoter(tx.ID, make(ObservedTxs, 0)),
		pool: Pool{
			Asset:        common.BNBAsset,
			BalanceRune:  sdk.NewUint(200),
			BalanceAsset: sdk.NewUint(300),
		},
		yggExists: true,
	}
	txOutStore := NewTxStoreDummy()

	vaultMgr := NewVaultMgrDummy()
	handler := NewObservedTxInHandler(keeper, txOutStore, w.poolAddrMgr, w.validatorMgr, vaultMgr)

	c.Assert(err, IsNil)
	msg := NewMsgObservedTxIn(txs, keeper.nas[0].NodeAddress)
	result := handler.handle(ctx, msg, ver)
	c.Assert(result.IsOK(), Equals, true)
	c.Check(txOutStore.GetOutboundItems(), HasLen, 1)
	c.Check(keeper.observing, HasLen, 1)
	c.Check(keeper.height.Equal(sdk.NewUint(12)), Equals, true)
	c.Check(keeper.chains, HasLen, 1)
	c.Check(keeper.chains[0].Equals(common.BNBChain), Equals, true)
}
