package thorchain

import (
	"encoding/json"

	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"

	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/constants"
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
	sErr := handler.validate(ctx, msgOutboundTx, ver)
	c.Assert(sErr, IsNil)

	// invalid version
	sErr = handler.validate(ctx, msgOutboundTx, semver.Version{})
	c.Assert(sErr, Equals, errBadVersion)

	// invalid msg
	msgOutboundTx = MsgOutboundTx{}
	sErr = handler.validate(ctx, msgOutboundTx, ver)
	c.Assert(sErr, NotNil)

	// not signed observer
	msgOutboundTx = NewMsgOutboundTx(tx, tx.Tx.ID, GetRandomBech32Addr())
	sErr = handler.validate(ctx, msgOutboundTx, ver)
	c.Assert(sErr.Code(), Equals, sdk.CodeUnauthorized)
}

type outboundTxHandlerTestHelper struct {
	ctx           sdk.Context
	pool          Pool
	version       semver.Version
	keeper        *outboundTxHandlerKeeperHelper
	asgardVault   Vault
	yggVault      Vault
	constAccessor constants.ConstantValues
	nodeAccount   NodeAccount
	inboundTx     ObservedTx
	toi           *TxOutItem
	event         Event
}

type outboundTxHandlerKeeperHelper struct {
	Keeper
	observeTxVoterErrHash common.TxID
	failGetPendingEvent   common.TxID
	errGetTxOut           bool
	errGetNodeAccount     bool
	errGetPool            bool
	errSetPool            bool
	errSetNodeAccount     bool
}

func newOutboundTxHandlerKeeperHelper(keeper Keeper) *outboundTxHandlerKeeperHelper {
	return &outboundTxHandlerKeeperHelper{
		Keeper:                keeper,
		observeTxVoterErrHash: GetRandomTxHash(),
		failGetPendingEvent:   GetRandomTxHash(),
	}
}

func (k *outboundTxHandlerKeeperHelper) GetObservedTxVoter(ctx sdk.Context, hash common.TxID) (ObservedTxVoter, error) {
	if hash.Equals(k.observeTxVoterErrHash) {
		return ObservedTxVoter{}, kaboom
	}
	return k.Keeper.GetObservedTxVoter(ctx, hash)
}
func (k *outboundTxHandlerKeeperHelper) GetPendingEventID(ctx sdk.Context, hash common.TxID) ([]int64, error) {
	if hash.Equals(k.failGetPendingEvent) {
		return nil, kaboom
	}
	return k.Keeper.GetPendingEventID(ctx, hash)
}
func (k *outboundTxHandlerKeeperHelper) GetTxOut(ctx sdk.Context, height int64) (*TxOut, error) {
	if k.errGetTxOut {
		return nil, kaboom
	}
	return k.Keeper.GetTxOut(ctx, height)
}
func (k *outboundTxHandlerKeeperHelper) GetNodeAccount(ctx sdk.Context, addr sdk.AccAddress) (NodeAccount, error) {
	if k.errGetNodeAccount {
		return NodeAccount{}, kaboom
	}
	return k.Keeper.GetNodeAccount(ctx, addr)
}
func (k *outboundTxHandlerKeeperHelper) GetPool(ctx sdk.Context, asset common.Asset) (Pool, error) {
	if k.errGetPool {
		return NewPool(), kaboom
	}
	return k.Keeper.GetPool(ctx, asset)
}
func (k *outboundTxHandlerKeeperHelper) SetPool(ctx sdk.Context, pool Pool) error {
	if k.errSetPool {
		return kaboom
	}
	return k.Keeper.SetPool(ctx, pool)
}

func (k *outboundTxHandlerKeeperHelper) SetNodeAccount(ctx sdk.Context, na NodeAccount) error {
	if k.errSetNodeAccount {
		return kaboom
	}
	return k.Keeper.SetNodeAccount(ctx, na)
}

// newOutboundTxHandlerTestHelper setup all the basic condition to test OutboundTxHandler
func newOutboundTxHandlerTestHelper(c *C) outboundTxHandlerTestHelper {
	ctx, k := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(1023)
	pool := NewPool()
	pool.Asset = common.BNBAsset
	pool.BalanceAsset = sdk.NewUint(100 * common.One)
	pool.BalanceRune = sdk.NewUint(100 * common.One)

	version := semver.MustParse("0.1.0")
	asgardVault := GetRandomVault()
	addr, err := asgardVault.PubKey.GetAddress(common.BNBChain)
	yggVault := GetRandomVault()
	c.Assert(err, IsNil)

	tx := NewObservedTx(common.Tx{
		ID:          GetRandomTxHash(),
		Chain:       common.BNBChain,
		Coins:       common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(1*common.One))},
		Memo:        "swap:RUNE-A1F",
		FromAddress: GetRandomBNBAddress(),
		ToAddress:   addr,
		Gas:         common.BNBGasFeeSingleton,
	}, 12, GetRandomPubKey())

	voter := NewObservedTxVoter(tx.Tx.ID, make(ObservedTxs, 0))
	keeper := newOutboundTxHandlerKeeperHelper(k)
	voter.Height = ctx.BlockHeight()
	keeper.SetObservedTxVoter(ctx, voter)
	nodeAccount := GetRandomNodeAccount(NodeActive)
	nodeAccount.Bond = sdk.NewUint(100 * common.One)
	c.Assert(keeper.SetNodeAccount(ctx, nodeAccount), IsNil)
	nodeAccount1 := GetRandomNodeAccount(NodeActive)
	nodeAccount1.Bond = sdk.NewUint(100 * common.One)

	c.Assert(keeper.SetPool(ctx, pool), IsNil)

	txOutStorage := NewTxOutStorageV1(keeper)
	constAccessor := constants.GetConstantValues(version)
	txOutStorage.NewBlock(ctx.BlockHeight(), constAccessor)
	toi := &TxOutItem{
		Chain:       common.BNBChain,
		ToAddress:   tx.Tx.FromAddress,
		VaultPubKey: yggVault.PubKey,
		Coin:        common.NewCoin(common.RuneAsset(), sdk.NewUint(common.One)),
		Memo:        NewOutboundMemo(tx.Tx.ID).String(),
		InHash:      tx.Tx.ID,
	}
	result, err := txOutStorage.TryAddTxOutItem(ctx, toi)
	txOutStorage.CommitBlock(ctx)

	swapEvent := NewEventSwap(common.BNBAsset, sdk.NewUint(common.One), sdk.NewUint(common.One), sdk.NewUint(common.One))
	buf, err := json.Marshal(swapEvent)
	c.Assert(err, IsNil)
	e := NewEvent(swapEvent.Type(), ctx.BlockHeight(), tx.Tx, buf, EventPending)
	c.Assert(keeper.UpsertEvent(ctx, e), IsNil)
	c.Assert(err, IsNil)
	c.Assert(result, Equals, true)
	return outboundTxHandlerTestHelper{
		ctx:           ctx,
		pool:          pool,
		version:       version,
		keeper:        keeper,
		asgardVault:   asgardVault,
		yggVault:      yggVault,
		nodeAccount:   nodeAccount,
		inboundTx:     tx,
		toi:           toi,
		constAccessor: constAccessor,
	}
}

func (s *HandlerOutboundTxSuite) TestOutboundTxHandlerShouldUpdateTxOut(c *C) {
	helper := newOutboundTxHandlerTestHelper(c)
	handler := NewOutboundTxHandler(helper.keeper)

	// invalid message should return an error
	noopMsg := NewMsgNoOp(helper.nodeAccount.NodeAddress)
	c.Assert(handler.Run(helper.ctx, noopMsg, helper.version, helper.constAccessor).Code, Equals, CodeInvalidMessage)

	fromAddr, err := helper.asgardVault.PubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	tx := NewObservedTx(common.Tx{
		ID:    GetRandomTxHash(),
		Chain: common.BNBChain,
		Coins: common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.NewUint(2*common.One)),
			common.NewCoin(common.BNBAsset, sdk.NewUint(common.One)),
		},
		Memo:        NewOutboundMemo(helper.inboundTx.Tx.ID).String(),
		FromAddress: fromAddr,
		ToAddress:   helper.inboundTx.Tx.FromAddress,
		Gas:         common.BNBGasFeeSingleton,
	}, helper.ctx.BlockHeight(), GetRandomPubKey())

	// if the version is lower than expected, it should return an error
	outMsg := NewMsgOutboundTx(tx, tx.Tx.ID, helper.nodeAccount.NodeAddress)
	c.Assert(handler.Run(helper.ctx, outMsg, semver.MustParse("0.0.1"), helper.constAccessor).Code, Equals, CodeBadVersion)

	// create a outbound tx with invalid observer account
	outMsg = NewMsgOutboundTx(tx, tx.Tx.ID, GetRandomNodeAccount(NodeActive).NodeAddress)
	// if the version is higher than expected, it should run as normal, because it should maintain backward compatibility
	c.Assert(handler.Run(helper.ctx, outMsg, semver.MustParse("0.2.0"), helper.constAccessor).Code, Equals, sdk.CodeUnauthorized)

	// fail to get observed TxVoter should result in an error
	outMsg = NewMsgOutboundTx(tx, helper.keeper.observeTxVoterErrHash, helper.nodeAccount.NodeAddress)
	c.Assert(handler.Run(helper.ctx, outMsg, semver.MustParse("0.1.0"), helper.constAccessor).Code, Equals, sdk.CodeInternal)

	// fail to complete events should result in an error
	outMsg = NewMsgOutboundTx(tx, helper.keeper.failGetPendingEvent, helper.nodeAccount.NodeAddress)
	c.Assert(handler.Run(helper.ctx, outMsg, semver.MustParse("0.1.0"), helper.constAccessor).Code, Equals, sdk.CodeInternal)

	// fail to get txout should result in an error
	helper.keeper.errGetTxOut = true
	outMsg = NewMsgOutboundTx(tx, tx.Tx.ID, helper.nodeAccount.NodeAddress)
	c.Assert(handler.Run(helper.ctx, outMsg, semver.MustParse("0.1.0"), helper.constAccessor).Code, Equals, sdk.CodeUnknownRequest)
	helper.keeper.errGetTxOut = false

	// fail to get node account should result in an error
	helper.keeper.errGetNodeAccount = true
	outMsg = NewMsgOutboundTx(tx, helper.inboundTx.Tx.ID, helper.nodeAccount.NodeAddress)
	c.Assert(handler.Run(helper.ctx, outMsg, semver.MustParse("0.1.0"), helper.constAccessor).Code, Equals, sdk.CodeInternal)
	helper.keeper.errGetNodeAccount = false

	// fail to get pool should result in an error
	helper.keeper.errGetPool = true
	outMsg = NewMsgOutboundTx(tx, helper.inboundTx.Tx.ID, helper.nodeAccount.NodeAddress)
	c.Assert(handler.Run(helper.ctx, outMsg, semver.MustParse("0.1.0"), helper.constAccessor).Code, Equals, sdk.CodeInternal)
	helper.keeper.errGetPool = false

	// fail to set pool should result in an error
	helper.keeper.errSetPool = true
	outMsg = NewMsgOutboundTx(tx, helper.inboundTx.Tx.ID, helper.nodeAccount.NodeAddress)
	c.Assert(handler.Run(helper.ctx, outMsg, semver.MustParse("0.1.0"), helper.constAccessor).Code, Equals, sdk.CodeInternal)
	helper.keeper.errSetPool = false

	// fail to set node account should result in an error
	helper.keeper.errSetNodeAccount = true
	outMsg = NewMsgOutboundTx(tx, helper.inboundTx.Tx.ID, helper.nodeAccount.NodeAddress)
	c.Assert(handler.Run(helper.ctx, outMsg, semver.MustParse("0.1.0"), helper.constAccessor).Code, Equals, sdk.CodeInternal)
	helper.keeper.errSetNodeAccount = false

	// valid outbound message, no event, no txout
	outMsg = NewMsgOutboundTx(tx, tx.Tx.ID, helper.nodeAccount.NodeAddress)
	c.Assert(handler.Run(helper.ctx, outMsg, semver.MustParse("0.1.0"), helper.constAccessor).Code, Equals, sdk.CodeOK)

}
func (s *HandlerOutboundTxSuite) TestOutboundTxNormalCase(c *C) {
	helper := newOutboundTxHandlerTestHelper(c)
	handler := NewOutboundTxHandler(helper.keeper)

	fromAddr, err := helper.asgardVault.PubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	tx := NewObservedTx(common.Tx{
		ID:    GetRandomTxHash(),
		Chain: common.BNBChain,
		Coins: common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.NewUint(common.One)),
		},
		Memo:        NewOutboundMemo(helper.inboundTx.Tx.ID).String(),
		FromAddress: fromAddr,
		ToAddress:   helper.inboundTx.Tx.FromAddress,
		Gas:         common.BNBGasFeeSingleton,
	}, helper.ctx.BlockHeight(), GetRandomPubKey())
	// valid outbound message, with event, with txout
	outMsg := NewMsgOutboundTx(tx, helper.inboundTx.Tx.ID, helper.nodeAccount.NodeAddress)
	c.Assert(handler.Run(helper.ctx, outMsg, semver.MustParse("0.1.0"), helper.constAccessor).Code, Equals, sdk.CodeOK)
	// event should set to complete
	ev, err := helper.keeper.GetEvent(helper.ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(ev.Status, Equals, EventSuccess)
	// txout should had been complete

	txOut, err := helper.keeper.GetTxOut(helper.ctx, helper.ctx.BlockHeight())
	c.Assert(err, IsNil)
	c.Assert(txOut.TxArray[0].OutHash.IsEmpty(), Equals, false)
}
func (s *HandlerOutboundTxSuite) TestOuboundTxHandlerSendExtraFundShouldBeSlashed(c *C) {
	helper := newOutboundTxHandlerTestHelper(c)
	handler := NewOutboundTxHandler(helper.keeper)
	fromAddr, err := helper.asgardVault.PubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	tx := NewObservedTx(common.Tx{
		ID:    GetRandomTxHash(),
		Chain: common.BNBChain,
		Coins: common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.NewUint(2*common.One)),
		},
		Memo:        NewOutboundMemo(helper.inboundTx.Tx.ID).String(),
		FromAddress: fromAddr,
		ToAddress:   helper.inboundTx.Tx.FromAddress,
		Gas:         common.BNBGasFeeSingleton,
	}, helper.ctx.BlockHeight(), helper.nodeAccount.PubKeySet.Secp256k1)
	expectedBond := helper.nodeAccount.Bond.Sub(sdk.NewUint(common.One).MulUint64(3).QuoUint64(2))
	// valid outbound message, with event, with txout
	outMsg := NewMsgOutboundTx(tx, helper.inboundTx.Tx.ID, helper.nodeAccount.NodeAddress)
	c.Assert(handler.Run(helper.ctx, outMsg, semver.MustParse("0.1.0"), helper.constAccessor).Code, Equals, sdk.CodeOK)
	na, err := helper.keeper.GetNodeAccount(helper.ctx, helper.nodeAccount.NodeAddress)
	c.Assert(na.Bond.Equal(expectedBond), Equals, true)
}

func (s *HandlerOutboundTxSuite) TestOutboundTxHandlerSendAdditionalCoinsShouldBeSlashed(c *C) {
	helper := newOutboundTxHandlerTestHelper(c)
	handler := NewOutboundTxHandler(helper.keeper)
	fromAddr, err := helper.asgardVault.PubKey.GetAddress(common.BNBChain)
	c.Assert(err, IsNil)
	tx := NewObservedTx(common.Tx{
		ID:    GetRandomTxHash(),
		Chain: common.BNBChain,
		Coins: common.Coins{
			common.NewCoin(common.RuneAsset(), sdk.NewUint(1*common.One)),
			common.NewCoin(common.BNBAsset, sdk.NewUint(1*common.One)),
		},
		Memo:        NewOutboundMemo(helper.inboundTx.Tx.ID).String(),
		FromAddress: fromAddr,
		ToAddress:   helper.inboundTx.Tx.FromAddress,
		Gas:         common.BNBGasFeeSingleton,
	}, helper.ctx.BlockHeight(), helper.nodeAccount.PubKeySet.Secp256k1)
	expectedBond := helper.nodeAccount.Bond.Sub(sdk.NewUint(common.One).MulUint64(3).QuoUint64(2))
	// slash one BNB
	outMsg := NewMsgOutboundTx(tx, helper.inboundTx.Tx.ID, helper.nodeAccount.NodeAddress)
	c.Assert(handler.Run(helper.ctx, outMsg, semver.MustParse("0.1.0"), helper.constAccessor).Code, Equals, sdk.CodeOK)
	na, err := helper.keeper.GetNodeAccount(helper.ctx, helper.nodeAccount.NodeAddress)
	c.Assert(na.Bond.Equal(expectedBond), Equals, true)
}
