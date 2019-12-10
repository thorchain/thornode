
package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/common"
	. "gopkg.in/check.v1"
)

type HandlerYggdrasilSuite struct{}

var _ = Suite(&HandlerYggdrasilSuite{})

func (s *HandlerYggdrasilSuite) TestValidate(c *C) {
	w := getHandlerTestWrapper(c, 1, true, true)
	handler := NewYggdrasilHandler(w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr)

	// happy path
	ver := semver.MustParse("0.1.0")
	pubKey := GetRandomPubKey()
	coins := common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One))}
	txID := GetRandomTxHash()
	signer := w.activeNodeAccount.NodeAddress
	msg := NewMsgYggdrasil(pubKey, true, coins, txID, signer)
	err := handler.validate(w.ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(w.ctx, msg, semver.Version{})
	c.Assert(err, Equals, badVersion)

	// invalid msg
	msg = MsgYggdrasil{}
	err = handler.validate(w.ctx, msg, ver)
	c.Assert(err, NotNil)

	// not authorized, not signed by active node account
	signer = GetRandomBech32Addr()
	msg = NewMsgYggdrasil(pubKey, true, coins, txID, signer)
	err = handler.validate(w.ctx, msg, ver)
	c.Assert(err, Equals, notAuthorized)
}

func (s *HandlerYggdrasilSuite) TestHandle(c *C) {
	w := getHandlerTestWrapper(c, 1, true, true)
	handler := NewYggdrasilHandler(w.keeper, w.txOutStore, w.poolAddrMgr, w.validatorMgr)
	acc2 := GetRandomNodeAccount(NodeActive)
	acc2.Bond = sdk.NewUint(100 * common.One)
	c.Assert(w.keeper.SetNodeAccount(w.ctx, acc2), IsNil)
	ygg := NewYggdrasil(acc2.NodePubKey.Secp256k1)
	c.Assert(w.keeper.SetYggdrasil(w.ctx, ygg), IsNil)

	// check yggdrasil balance on add funds
	ver := semver.MustParse("0.1.0")
	pubKey := w.activeNodeAccount.NodePubKey.Secp256k1
	coins := common.Coins{common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One))}
	txID := GetRandomTxHash()
	signer := w.activeNodeAccount.NodeAddress
	msg := NewMsgYggdrasil(pubKey, true, coins, txID, signer)
	result := handler.handle(w.ctx, msg, ver)
	c.Assert(result.Code, Equals, sdk.CodeOK)

	ygg, err := handler.keeper.GetYggdrasil(w.ctx, msg.PubKey)
	c.Assert(err, IsNil)
	coin := ygg.GetCoin(common.BNBAsset)
	c.Check(coin.Amount.Uint64(), Equals, sdk.NewUint(100*common.One).Uint64(), Commentf("%d vs %d", coin.Amount.Uint64(), sdk.NewUint(100*common.One).Uint64()))

	// check yggdrasil balance on sub funds
	msg = NewMsgYggdrasil(pubKey, false, coins, txID, signer)
	result = handler.handle(w.ctx, msg, ver)
	c.Assert(result.Code, Equals, sdk.CodeOK)

	ygg, err = handler.keeper.GetYggdrasil(w.ctx, msg.PubKey)
	c.Assert(err, IsNil)
	coin = ygg.GetCoin(common.BNBAsset)
	c.Check(coin.Amount.Uint64(), Equals, sdk.NewUint(0*common.One).Uint64(), Commentf("%d vs %d", coin.Amount.Uint64(), sdk.NewUint(0*common.One).Uint64()))

	// trigger Ragnarok step2
	handler.validatorMgr.Meta().Ragnarok = true
	result = handler.handle(w.ctx, msg, ver)
	c.Assert(result.Code, Equals, sdk.CodeOK)
}
