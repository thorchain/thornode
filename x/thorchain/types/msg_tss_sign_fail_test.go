package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type MsgTssKeysignFailSuite struct{}

var _ = Suite(&MsgTssKeysignFailSuite{})

func (s MsgTssKeysignFailSuite) TestMsgTssKeysignFail(c *C) {
	b := common.Blame{
		FailReason: "fail to TSS sign",
		BlameNodes: common.PubKeys{
			GetRandomPubKey(),
			GetRandomPubKey(),
		},
	}
	coins := common.Coins{
		common.NewCoin(common.RuneAsset(), sdk.NewUint(100)),
	}
	msg := NewMsgTssKeysignFail(1, b, "hello", coins, GetRandomBech32Addr())
	c.Check(msg.Type(), Equals, "set_tss_keysign_fail")
	EnsureMsgBasicCorrect(msg, c)
	c.Check(NewMsgTssKeysignFail(1, common.EmptyBlame, "hello", coins, GetRandomBech32Addr()), NotNil)
	c.Check(NewMsgTssKeysignFail(1, b, "", coins, GetRandomBech32Addr()), NotNil)
	c.Check(NewMsgTssKeysignFail(1, b, "hello", common.Coins{}, GetRandomBech32Addr()), NotNil)
	c.Check(NewMsgTssKeysignFail(1, b, "hello", common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(100)),
		common.NewCoin(common.EmptyAsset, sdk.ZeroUint()),
	}, GetRandomBech32Addr()), NotNil)
	c.Check(NewMsgTssKeysignFail(1, b, "hello", coins, sdk.AccAddress{}), NotNil)
}
