package bitcoin

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/x/thorchain"
)

type BitcoinUTXOAccessor struct{}

var _ = Suite(
	&BitcoinUTXOAccessor{},
)

func (s *BitcoinUTXOAccessor) TestNewUTXOAccessor(c *C) {
	memStorage := storage.NewMemStorage()
	db, err := leveldb.Open(memStorage, nil)
	c.Assert(err, IsNil)
	utxoAccessor, err := NewLevelDBUTXOAccessor(db)
	c.Assert(err, IsNil)
	c.Assert(utxoAccessor, NotNil)
}

func (s *BitcoinUTXOAccessor) TestUTXOAccessor(c *C) {
	memStorage := storage.NewMemStorage()
	db, err := leveldb.Open(memStorage, nil)
	c.Assert(err, IsNil)
	utxoAccessor, err := NewLevelDBUTXOAccessor(db)
	c.Assert(err, IsNil)
	c.Assert(utxoAccessor, NotNil)

	txID, err := chainhash.NewHashFromStr("31f8699ce9028e9cd37f8a6d58a79e614a96e3fdd0f58be5fc36d2d95484716f")
	c.Assert(err, IsNil)
	pkey := thorchain.GetRandomPubKey()
	utxo := NewUnspentTransactionOutput(*txID, 0, 1, 10, pkey)
	err = utxoAccessor.AddUTXO(utxo)
	c.Assert(err, IsNil)

	utxos, err := utxoAccessor.GetUTXOs(pkey)
	c.Assert(err, IsNil)
	c.Assert(len(utxos), Equals, 1)
	c.Assert(utxos[0].GetKey(), Equals, "31f8699ce9028e9cd37f8a6d58a79e614a96e3fdd0f58be5fc36d2d95484716f:0")
	c.Assert(utxos[0].Value, Equals, float64(1))
	c.Assert(utxos[0].BlockHeight, Equals, int64(10))

	// add again and check still unique
	err = utxoAccessor.AddUTXO(utxo)
	c.Assert(err, IsNil)

	utxos, err = utxoAccessor.GetUTXOs(pkey)
	c.Assert(err, IsNil)
	c.Assert(len(utxos), Equals, 1)
	c.Assert(utxos[0].GetKey(), Equals, "31f8699ce9028e9cd37f8a6d58a79e614a96e3fdd0f58be5fc36d2d95484716f:0")

	// add another one
	txID, err = chainhash.NewHashFromStr("24ed2d26fd5d4e0e8fa86633e40faf1bdfc8d1903b1cd02855286312d48818a2")
	c.Assert(err, IsNil)
	utxo = NewUnspentTransactionOutput(*txID, 1, 2, 1234, pkey)
	err = utxoAccessor.AddUTXO(utxo)
	c.Assert(err, IsNil)

	utxos, err = utxoAccessor.GetUTXOs(pkey)
	c.Assert(err, IsNil)
	c.Assert(len(utxos), Equals, 2)
	c.Assert(utxos[0].GetKey(), Equals, "24ed2d26fd5d4e0e8fa86633e40faf1bdfc8d1903b1cd02855286312d48818a2:1")
	c.Assert(utxos[0].Value, Equals, float64(2))
	c.Assert(utxos[0].BlockHeight, Equals, int64(1234))
	c.Assert(utxos[1].GetKey(), Equals, "31f8699ce9028e9cd37f8a6d58a79e614a96e3fdd0f58be5fc36d2d95484716f:0")
	c.Assert(utxos[1].Value, Equals, float64(1))
	c.Assert(utxos[1].BlockHeight, Equals, int64(10))

	// delete one
	err = utxoAccessor.RemoveUTXO(utxo.GetKey())
	c.Assert(err, IsNil)

	utxos, err = utxoAccessor.GetUTXOs(pkey)
	c.Assert(err, IsNil)
	c.Assert(len(utxos), Equals, 1)
	c.Assert(utxos[0].GetKey(), Equals, "31f8699ce9028e9cd37f8a6d58a79e614a96e3fdd0f58be5fc36d2d95484716f:0")
	c.Assert(utxos[0].Value, Equals, float64(1))
	c.Assert(utxos[0].BlockHeight, Equals, int64(10))

	fee, vSize, err := utxoAccessor.GetTransactionFee()
	c.Assert(err, NotNil)
	c.Assert(fee, Equals, 0.0)
	c.Assert(vSize, Equals, int32(0))
	// upsert transaction fee
	c.Assert(utxoAccessor.UpsertTransactionFee(1.0, 1), IsNil)
	fee, vSize, err = utxoAccessor.GetTransactionFee()
	c.Assert(err, IsNil)
	c.Assert(fee, Equals, 1.0)
	c.Assert(vSize, Equals, int32(1))
}
