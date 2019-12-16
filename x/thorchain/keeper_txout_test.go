package thorchain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type KeeperTxOutSuite struct{}

var _ = Suite(&KeeperTxOutSuite{})

func (KeeperTxOutSuite) TestKeeperTxOut(c *C) {
	ctx, k := setupKeeperForTest(c)
	txOut := NewTxOut(1)
	txOutItem := &TxOutItem{
		Chain:       common.BNBChain,
		ToAddress:   GetRandomBNBAddress(),
		VaultPubKey: GetRandomPubKey(),
		SeqNo:       0,
		Coin:        common.NewCoin(common.BNBAsset, sdk.NewUint(100*common.One)),
		Memo:        "hello",
	}
	txOut.TxArray = append(txOut.TxArray, txOutItem)
	c.Assert(k.SetTxOut(ctx, txOut), IsNil)
	txOut1, err := k.GetTxOut(ctx, 1)
	c.Assert(err, IsNil)
	c.Assert(txOut1, NotNil)
	c.Assert(txOut1.Height, Equals, uint64(1))

	txOut2, err := k.GetTxOut(ctx, 100)
	c.Assert(err, IsNil)
	c.Assert(txOut2, NotNil)

	iter := k.GetTxOutIterator(ctx)
	defer iter.Close()
}
