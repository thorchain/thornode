package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type HandlerTssSuite struct{}

type TestTssValidKeepr struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestTssValidKeepr) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

var _ = Suite(&HandlerTssSuite{})

func (s *HandlerTssSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestTssValidKeepr{
		na: GetRandomNodeAccount(NodeActive),
	}
	txOutStore := NewTxStoreDummy()
	poolAddrMgr := NewPoolAddressDummyMgr()

	handler := NewTssHandler(keeper, txOutStore, poolAddrMgr)
	// happy path
	ver := semver.MustParse("0.1.0")
	pk := GetRandomPubKey()
	pks := []common.PubKey{
		GetRandomPubKey(), GetRandomPubKey(), GetRandomPubKey(),
	}
	msg := NewMsgTssPool(pks, pk, keeper.na.NodeAddress)
	err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// inactive node account
	keeper.na = GetRandomNodeAccount(NodeStandby)
	msg = NewMsgTssPool(pks, pk, keeper.na.NodeAddress)
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, Equals, notAuthorized)

	// invalid msg
	msg = MsgTssPool{}
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

type TestTssHandlerKeeper struct {
	KVStoreDummy
	active NodeAccounts
	tss    TssVoter
	chains common.Chains
}

func (s *TestTssHandlerKeeper) ListActiveNodeAccounts(_ sdk.Context) (NodeAccounts, error) {
	return s.active, nil
}

func (s *TestTssHandlerKeeper) GetTssVoter(_ sdk.Context, _ string) (TssVoter, error) {
	return s.tss, nil
}

func (s *TestTssHandlerKeeper) SetTssVoter(_ sdk.Context, voter TssVoter) {
	s.tss = voter
}

func (s *TestTssHandlerKeeper) GetChains(_ sdk.Context) (common.Chains, error) {
	return s.chains, nil
}

type TestTssPoolMgr struct {
	PoolAddressDummyMgr
	pks common.PoolPubKeys
}

func (p *TestTssPoolMgr) RotatePoolAddress(_ sdk.Context, pks common.PoolPubKeys, _ TxOutStore) {
	p.pks = pks
}

func (s *HandlerTssSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(12)
	ver := semver.MustParse("0.1.0")

	keeper := &TestTssHandlerKeeper{
		active: NodeAccounts{GetRandomNodeAccount(NodeActive)},
		chains: common.Chains{common.BNBChain},
		tss:    TssVoter{},
	}
	txOutStore := NewTxStoreDummy()
	poolAddrMgr := &TestTssPoolMgr{}

	handler := NewTssHandler(keeper, txOutStore, poolAddrMgr)
	// happy path
	pk := GetRandomPubKey()
	pks := []common.PubKey{
		GetRandomPubKey(), GetRandomPubKey(), GetRandomPubKey(),
	}
	msg := NewMsgTssPool(pks, pk, keeper.active[0].NodeAddress)
	result := handler.handle(ctx, msg, ver)
	c.Assert(result.IsOK(), Equals, true)
	c.Check(keeper.tss.Signers, HasLen, 1)
	c.Check(keeper.tss.BlockHeight, Equals, int64(12))
	c.Check(poolAddrMgr.pks.IsEmpty(), Equals, false)

	// running again doesn't rotate the pool again
	ctx = ctx.WithBlockHeight(14)
	result = handler.handle(ctx, msg, ver)
	c.Assert(result.IsOK(), Equals, true)
	c.Check(keeper.tss.BlockHeight, Equals, int64(12))
}
