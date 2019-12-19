package thorchain

import (
	"fmt"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type HandlerObservedTxOutSuite struct{}

type TestObservedTxOutValidateKeeper struct {
	KVStoreDummy
	isActive bool
}

func (k *TestObservedTxOutValidateKeeper) IsActiveObserver(ctx sdk.Context, signer sdk.AccAddress) bool {
	return k.isActive
}

var _ = Suite(&HandlerObservedTxOutSuite{})

func (s *HandlerObservedTxOutSuite) TestValidate(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	keeper := &TestObservedTxOutValidateKeeper{
		isActive: true,
	}

	handler := NewObservedTxOutHandler(keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr)

	// happy path
	ver := semver.MustParse("0.1.0")
	pk := GetRandomPubKey()
	txs := ObservedTxs{NewObservedTx(GetRandomTx(), sdk.NewUint(12), pk)}
	txs[0].Tx.FromAddress, err = pk.GetAddress(txs[0].Tx.Coins[0].Asset.Chain)
	c.Assert(err, IsNil)
	msg := NewMsgObservedTxOut(txs, GetRandomBech32Addr())
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// inactive node account
	keeper.isActive = false
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

func (s *HandlerObservedTxOutSuite) TestFailure(c *C) {
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	keeper := &TestObservedTxOutFailureKeeper{}
	txOutStore := NewTxStoreDummy()

	handler := NewObservedTxOutHandler(keeper, txOutStore, w.poolAddrMgr, w.validatorMgr)
	tx := NewObservedTx(GetRandomTx(), sdk.NewUint(12), GetRandomPubKey())
	nas := NodeAccounts{GetRandomNodeAccount(NodeActive)}

	err := handler.outboundFailure(ctx, tx, nas)
	c.Assert(err, IsNil)
}

type TestObservedTxOutHandleKeeper struct {
	KVStoreDummy
	nas        NodeAccounts
	na         NodeAccount
	voter      ObservedTxVoter
	yggExists  bool
	ygg        Vault
	height     sdk.Uint
	chains     common.Chains
	pool       Pool
	txOutStore TxOutStore
	observing  []sdk.AccAddress
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

func (k *TestObservedTxOutHandleKeeper) GetChains(_ sdk.Context) (common.Chains, error) {
	return k.chains, nil
}

func (k *TestObservedTxOutHandleKeeper) SetChains(_ sdk.Context, chains common.Chains) {
	k.chains = chains
}

func (k *TestObservedTxOutHandleKeeper) SetLastChainHeight(_ sdk.Context, _ common.Chain, height sdk.Uint) error {
	k.height = height
	return nil
}

func (k *TestObservedTxOutHandleKeeper) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return k.pool, nil
}

func (k *TestObservedTxOutHandleKeeper) AddIncompleteEvents(_ sdk.Context, evt Event) error {
	return nil
}

func (k *TestObservedTxOutHandleKeeper) GetTxOut(_ sdk.Context, _ uint64) (*TxOut, error) {
	return k.txOutStore.GetBlockOut(), nil
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

func (k *TestObservedTxOutHandleKeeper) GetLastEventID(_ sdk.Context) (int64, error) {
	return 0, nil
}

func (k *TestObservedTxOutHandleKeeper) GetIncompleteEvents(_ sdk.Context) (Events, error) {
	return nil, nil
}

func (s *HandlerObservedTxOutSuite) TestHandle(c *C) {
	var err error
	ctx, _ := setupKeeperForTest(c)
	w := getHandlerTestWrapper(c, 1, true, false)

	ver := semver.MustParse("0.1.0")
	tx := GetRandomTx()
	tx.Memo = fmt.Sprintf("OUTBOUND:%s", tx.ID)
	obTx := NewObservedTx(tx, sdk.NewUint(12), GetRandomPubKey())
	txs := ObservedTxs{obTx}
	pk := GetRandomPubKey()
	currentPool := w.poolAddrMgr.GetCurrentPoolAddresses().Current.GetByChain(tx.Chain)
	txs[0].Tx.FromAddress, err = currentPool.GetAddress()
	c.Assert(err, IsNil)

	txOutStore := NewTxStoreDummy()
	keeper := &TestObservedTxOutHandleKeeper{
		nas:   NodeAccounts{GetRandomNodeAccount(NodeActive)},
		voter: NewObservedTxVoter(tx.ID, make(ObservedTxs, 0)),
		pool: Pool{
			Asset:        common.BNBAsset,
			BalanceRune:  sdk.NewUint(200),
			BalanceAsset: sdk.NewUint(300),
		},
		yggExists: true,
		ygg: Vault{
			PubKey: pk,
			Coins: common.Coins{
				common.NewCoin(common.RuneAsset(), sdk.NewUint(500)),
				common.NewCoin(common.BNBAsset, sdk.NewUint(200)),
			},
			Type: YggdrasilVault,
		},
		txOutStore: txOutStore,
	}

	handler := NewObservedTxOutHandler(keeper, txOutStore, w.poolAddrMgr, w.validatorMgr)

	c.Assert(err, IsNil)
	msg := NewMsgObservedTxOut(txs, keeper.nas[0].NodeAddress)
	result := handler.handle(ctx, msg, ver)
	c.Assert(result.IsOK(), Equals, true)
	c.Check(txOutStore.GetOutboundItems(), HasLen, 0)
	c.Check(keeper.observing, HasLen, 1)
}
