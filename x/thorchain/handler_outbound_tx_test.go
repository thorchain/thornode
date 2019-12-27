package thorchain

import (
	"errors"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
	"gitlab.com/thorchain/thornode/x/thorchain/types"

	. "gopkg.in/check.v1"
)

type HandlerOutboundTxSuite struct{}

type TestOutboundTxKeeper struct {
	KVStoreDummy
	activeNodeAccount NodeAccount
	vault             Vault
}

// IsActiveObserver see whether it is an active observer
func (k *TestOutboundTxKeeper) IsActiveObserver(_ sdk.Context, addr sdk.AccAddress) bool {
	return k.activeNodeAccount.NodeAddress.Equals(addr)
}

var _ = Suite(&HandlerOutboundTxSuite{})

func (s *HandlerOutboundTxSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestOutboundTxKeeper{
		activeNodeAccount: GetRandomNodeAccount(NodeActive),
		vault:             GetRandomVault(),
	}

	handler := NewOutboundTxHandler(keeper)

	addr, err := keeper.vault.PubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)

	ver := semver.MustParse("0.1.0")

	tx := NewObservedTx(common.Tx{
		ID:          GetRandomTxHash(),
		Chain:       common.BNBChain,
		Coins:       common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(1*common.One))},
		Memo:        "",
		FromAddress: GetRandomBNBAddress(),
		ToAddress:   addr,
		Gas:         common.BNBGasFeeSingleton,
	}, 12, GetRandomPubKey())

	msgOutboundTx := NewMsgOutboundTx(tx, tx.Tx.ID, keeper.activeNodeAccount.NodeAddress)
	err = handler.validate(ctx, msgOutboundTx, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msgOutboundTx, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// invalid msg
	msgOutboundTx = MsgOutboundTx{}
	err = handler.validate(ctx, msgOutboundTx, ver)
	c.Assert(err, NotNil)

	// not signed observer
	msgOutboundTx = NewMsgOutboundTx(tx, tx.Tx.ID, GetRandomBech32Addr())
	err = handler.validate(ctx, msgOutboundTx, ver)
	c.Assert(err, Equals, notAuthorized)
}

type TestOutboundTxHandleKeeper struct {
	KVStoreDummy
	asgardVault       Vault
	activeNodeAccount NodeAccount
	voter             ObservedTxVoter
	event             Event
	pool              Pool
	txOut             TxOut
	height            int64
	observing         []sdk.AccAddress
	chains            common.Chains
}

func (k *TestOutboundTxHandleKeeper) GetChains(_ sdk.Context) (common.Chains, error) {
	return k.chains, nil
}

func (k *TestOutboundTxHandleKeeper) SetChains(_ sdk.Context, chains common.Chains) {
	k.chains = chains
}

func (k *TestOutboundTxHandleKeeper) SetLastChainHeight(_ sdk.Context, _ common.Chain, height int64) error {
	k.height = height
	return nil
}

func (k *TestOutboundTxHandleKeeper) VaultExists(_ sdk.Context, _ common.PubKey) bool {
	return !k.asgardVault.IsEmpty()
}

func (k *TestOutboundTxHandleKeeper) GetVault(_ sdk.Context, _ common.PubKey) (Vault, error) {
	return k.asgardVault, nil
}

func (k *TestOutboundTxHandleKeeper) SetVault(_ sdk.Context, vault Vault) error {
	k.asgardVault = vault
	return nil
}

func (k *TestOutboundTxHandleKeeper) GetObservedTxVoter(_ sdk.Context, _ common.TxID) (ObservedTxVoter, error) {
	return k.voter, nil
}

func (k *TestOutboundTxHandleKeeper) SetObservedTxVoter(_ sdk.Context, voter ObservedTxVoter) {
	k.voter = voter
}

func (k *TestOutboundTxHandleKeeper) GetPendingEventID(_ sdk.Context, _ common.TxID) ([]int64, error) {
	return []int64{k.event.ID}, nil
}

func (k *TestOutboundTxHandleKeeper) GetEvent(_ sdk.Context, eventID int64) (Event, error) {
	if eventID == k.event.ID {
		return k.event, nil
	}
	return Event{}, kaboom
}

func (k *TestOutboundTxHandleKeeper) UpsertEvent(_ sdk.Context, event Event) error {
	k.event = event
	return nil
}

// IsActiveObserver see whether it is an active observer
func (k *TestOutboundTxHandleKeeper) IsActiveObserver(_ sdk.Context, addr sdk.AccAddress) bool {
	return k.activeNodeAccount.NodeAddress.Equals(addr)
}

func (k *TestOutboundTxHandleKeeper) ListActiveNodeAccounts(_ sdk.Context) (NodeAccounts, error) {
	return NodeAccounts{k.activeNodeAccount}, nil
}

func (k *TestOutboundTxHandleKeeper) GetNodeAccount(_ sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	if k.activeNodeAccount.NodeAddress.Equals(addr) {
		return k.activeNodeAccount, nil
	}
	return NodeAccount{}, errors.New("not exist")
}

func (k *TestOutboundTxHandleKeeper) GetVaultData(_ sdk.Context) (VaultData, error) {
	return NewVaultData(), nil
}

func (k *TestOutboundTxHandleKeeper) SetVaultData(_ sdk.Context, _ VaultData) error {
	return nil
}

func (k *TestOutboundTxHandleKeeper) GetPool(_ sdk.Context, _ common.Asset) (Pool, error) {
	return k.pool, nil
}

func (k *TestOutboundTxHandleKeeper) SetPool(_ sdk.Context, pool Pool) error {
	k.pool = pool
	return nil
}

func (k *TestOutboundTxHandleKeeper) GetTxOut(_ sdk.Context, _ uint64) (*TxOut, error) {
	return &k.txOut, nil
}

func (k *TestOutboundTxHandleKeeper) SetTxOut(_ sdk.Context, _ *TxOut) error {
	return nil
}

func (k *TestOutboundTxHandleKeeper) AddIncompleteEvents(_ sdk.Context, evt Event) error {
	return nil
}

func (k *TestOutboundTxHandleKeeper) AddObservingAddresses(_ sdk.Context, addrs []sdk.AccAddress) error {
	k.observing = addrs
	return nil
}

func (s *HandlerOutboundTxSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestOutboundTxHandleKeeper{
		activeNodeAccount: GetRandomNodeAccount(NodeActive),
		asgardVault:       GetRandomVault(),
	}

	ver := semver.MustParse("0.1.0")

	pool := NewPool()
	pool.Asset = common.BNBAsset
	pool.BalanceAsset = sdk.NewUint(100 * common.One)
	pool.BalanceRune = sdk.NewUint(100 * common.One)
	c.Assert(keeper.SetPool(ctx, pool), IsNil)

	constAccessor := constants.GetConstantValues(ver)
	vaultMgr := NewVaultMgrDummy()
	txOutStore := NewTxStoreDummy()
	validatorMgr := NewValidatorMgr(keeper, txOutStore, vaultMgr)

	handler := NewOutboundTxHandler(keeper)

	addr, err := keeper.asgardVault.PubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)

	tx := NewObservedTx(common.Tx{
		ID:          GetRandomTxHash(),
		Chain:       common.BNBChain,
		Coins:       common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(1*common.One))},
		Memo:        "",
		FromAddress: GetRandomBNBAddress(),
		ToAddress:   addr,
		Gas:         common.BNBGasFeeSingleton,
	}, 12, GetRandomPubKey())

	voter := NewObservedTxVoter(tx.Tx.ID, make(ObservedTxs, 0))
	keeper.SetObservedTxVoter(ctx, voter)

	ygg := NewVault(ctx.BlockHeight(), ActiveVault, YggdrasilVault, keeper.asgardVault.PubKey)
	ygg.Coins = common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(500*common.One)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(400*common.One)),
	}
	c.Assert(keeper.SetVault(ctx, ygg), IsNil)

	tx.ObservedPubKey = keeper.asgardVault.PubKey
	tx.Tx.FromAddress = addr
	tx.Tx.Coins = common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(200*common.One)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(200*common.One)),
	}
	tx.Tx.ID = GetRandomTxHash()
	msgOutboundTxNormal := NewMsgOutboundTx(tx, tx.Tx.ID, keeper.activeNodeAccount.NodeAddress)
	result3 := handler.handle(ctx, msgOutboundTxNormal, ver)
	c.Assert(result3.Code, Equals, sdk.CodeOK, Commentf("%+v\n", result3))
	ygg, err = keeper.GetVault(ctx, keeper.asgardVault.PubKey)
	c.Assert(err, IsNil)
	c.Check(ygg.GetCoin(common.BNBAsset).Amount.Equal(sdk.NewUint(29999962500)), Equals, true) // 300 - Gas
	c.Check(ygg.GetCoin(common.BTCAsset).Amount.Equal(sdk.NewUint(200*common.One)), Equals, true)
	txOutStore.NewBlock(2, constAccessor)
	inTxID := GetRandomTxHash()

	txIn := types.NewObservedTx(
		common.Tx{
			ID:          inTxID,
			Chain:       common.BNBChain,
			Coins:       common.Coins{common.NewCoin(common.RuneAsset(), sdk.NewUint(1*common.One))},
			Memo:        "swap:BNB",
			FromAddress: GetRandomBNBAddress(),
			ToAddress:   addr,
			Gas:         common.BNBGasFeeSingleton,
		},
		1024,
		keeper.asgardVault.PubKey,
	)

	observedTxInHandler := NewObservedTxInHandler(keeper, txOutStore, validatorMgr, vaultMgr)
	msgObservedTxIn := NewMsgObservedTxIn(ObservedTxs{txIn}, keeper.activeNodeAccount.NodeAddress)
	result := observedTxInHandler.Run(ctx, msgObservedTxIn, ver, constAccessor)
	c.Assert(result.Code, Equals, sdk.CodeOK, Commentf("%s\n", result.Log))

	tx = NewObservedTx(common.Tx{
		ID:          GetRandomTxHash(),
		Chain:       common.BNBChain,
		Coins:       common.Coins{common.NewCoin(common.RuneAsset(), sdk.NewUint(1*common.One))},
		Memo:        "swap:BNB",
		FromAddress: GetRandomBNBAddress(),
		ToAddress:   GetRandomBNBAddress(),
		Gas:         common.BNBGasFeeSingleton,
	}, 12, GetRandomPubKey())

	outMsg := NewMsgOutboundTx(tx, inTxID, keeper.activeNodeAccount.NodeAddress)
	ctx = ctx.WithBlockHeight(2)
	result4 := handler.handle(ctx, outMsg, ver)
	c.Assert(result4.Code, Equals, sdk.CodeOK, Commentf("%+v\n", result4))
}
