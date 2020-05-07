package bitcoin

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/x/thorchain"
)

type BlockMetaTestSuite struct{}

var _ = Suite(
	&BlockMetaTestSuite{},
)

func (b *BlockMetaTestSuite) TestBlockMeta(c *C) {
	blockMeta := NewBlockMeta("00000000000000d9cba4b81d1f8fb5cecd54e4ec3104763ba937aa7692a86dc5",
		1722479,
		"00000000000000ca7a4633264b9989355e9709f9e9da19506b0f636cc435dc8f")
	c.Assert(blockMeta, NotNil)

	txID, err := chainhash.NewHashFromStr("31f8699ce9028e9cd37f8a6d58a79e614a96e3fdd0f58be5fc36d2d95484716f")
	c.Assert(err, IsNil)
	pkey := thorchain.GetRandomPubKey()
	utxo := NewUnspentTransactionOutput(*txID, 0, 1, 10, pkey)
	blockMeta.AddUTXO(utxo)

	utxos := blockMeta.GetUTXOs(pkey)
	c.Assert(len(utxos), Equals, 1)
	c.Assert(utxos[0].GetKey(), Equals, "31f8699ce9028e9cd37f8a6d58a79e614a96e3fdd0f58be5fc36d2d95484716f:0")
	c.Assert(utxos[0].Value, Equals, float64(1))
	c.Assert(utxos[0].BlockHeight, Equals, int64(10))

	// add again and check still unique
	blockMeta.AddUTXO(utxo)
	utxos = blockMeta.GetUTXOs(pkey)
	c.Assert(err, IsNil)
	c.Assert(len(utxos), Equals, 1)
	c.Assert(utxos[0].GetKey(), Equals, "31f8699ce9028e9cd37f8a6d58a79e614a96e3fdd0f58be5fc36d2d95484716f:0")

	// add another one
	txID, err = chainhash.NewHashFromStr("24ed2d26fd5d4e0e8fa86633e40faf1bdfc8d1903b1cd02855286312d48818a2")
	c.Assert(err, IsNil)
	utxo = NewUnspentTransactionOutput(*txID, 1, 2, 1234, pkey)
	blockMeta.AddUTXO(utxo)

	utxos = blockMeta.GetUTXOs(pkey)
	c.Assert(err, IsNil)
	c.Assert(len(utxos), Equals, 2)
	c.Assert(utxos[0].GetKey(), Equals, "31f8699ce9028e9cd37f8a6d58a79e614a96e3fdd0f58be5fc36d2d95484716f:0")
	c.Assert(utxos[0].Value, Equals, float64(1))
	c.Assert(utxos[0].BlockHeight, Equals, int64(10))
	c.Assert(utxos[1].GetKey(), Equals, "24ed2d26fd5d4e0e8fa86633e40faf1bdfc8d1903b1cd02855286312d48818a2:1")
	c.Assert(utxos[1].Value, Equals, float64(2))
	c.Assert(utxos[1].BlockHeight, Equals, int64(1234))

	// delete one
	blockMeta.RemoveUTXO(utxo.GetKey())

	utxos = blockMeta.GetUTXOs(pkey)
	c.Assert(err, IsNil)
	c.Assert(len(utxos), Equals, 1)
	c.Assert(utxos[0].GetKey(), Equals, "31f8699ce9028e9cd37f8a6d58a79e614a96e3fdd0f58be5fc36d2d95484716f:0")
	c.Assert(utxos[0].Value, Equals, float64(1))
	c.Assert(utxos[0].BlockHeight, Equals, int64(10))

	// mark as spent
	utxo = blockMeta.UnspentTransactionOutputs[0]
	c.Assert(utxo.Spent, Equals, false)
	blockMeta.SpendUTXO(utxo.GetKey())
	utxo = blockMeta.UnspentTransactionOutputs[0]
	c.Assert(utxo.Spent, Equals, true)

	// check getutxos dont return the spent one
	utxos = blockMeta.GetUTXOs(pkey)
	c.Assert(err, IsNil)
	c.Assert(len(utxos), Equals, 0)

	// mark as unspent
	utxo = blockMeta.UnspentTransactionOutputs[0]
	c.Assert(utxo.Spent, Equals, true)
	blockMeta.UnspendUTXO(utxo.GetKey())
	utxo = blockMeta.UnspentTransactionOutputs[0]
	c.Assert(utxo.Spent, Equals, false)

	// check getutxos return the unspent one
	utxos = blockMeta.GetUTXOs(pkey)
	c.Assert(err, IsNil)
	c.Assert(len(utxos), Equals, 1)
}
