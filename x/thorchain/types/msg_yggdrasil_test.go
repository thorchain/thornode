package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
	. "gopkg.in/check.v1"
)

type MsgYggdrasilSuite struct{}

var _ = Suite(&MsgYggdrasilSuite{})

func (s *MsgYggdrasilSuite) TestMsgYggdrasil(c *C) {
	txId := GetRandomTxHash()
	pk := GetRandomPubKey()
	coins := common.Coins{
		common.NewCoin(common.BNBAsset, sdk.NewUint(500*common.One)),
		common.NewCoin(common.BTCAsset, sdk.NewUint(400*common.One)),
	}
	signer := GetRandomBech32Addr()

	msg := NewMsgYggdrasil(pk, true, coins, txId, signer)
	c.Check(msg.PubKey.Equals(pk), Equals, true)
	c.Check(msg.AddFunds, Equals, true)
	c.Check(msg.Coins, HasLen, len(coins))
	c.Check(msg.RequestTxHash.Equals(txId), Equals, true)
	c.Check(msg.Signer.Equals(signer), Equals, true)
}
