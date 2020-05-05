package bitcoin

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/x/thorchain"
)

type BitcoinBlockMetaAccessorTestSuite struct{}

var _ = Suite(
	&BitcoinBlockMetaAccessorTestSuite{},
)

func (s *BitcoinBlockMetaAccessorTestSuite) TestNewBlockMetaAccessor(c *C) {
	memStorage := storage.NewMemStorage()
	db, err := leveldb.Open(memStorage, nil)
	c.Assert(err, IsNil)
	dbBlockMetaAccessor, err := NewLevelDBBlockMetaAccessor(db)
	c.Assert(err, IsNil)
	c.Assert(dbBlockMetaAccessor, NotNil)
}

func (s *BitcoinBlockMetaAccessorTestSuite) TestBlockMetaAccessor(c *C) {
	memStorage := storage.NewMemStorage()
	db, err := leveldb.Open(memStorage, nil)
	c.Assert(err, IsNil)
	blockMetaAccessor, err := NewLevelDBBlockMetaAccessor(db)
	c.Assert(err, IsNil)
	c.Assert(blockMetaAccessor, NotNil)

	blockMeta := NewBlockMeta("00000000000000d9cba4b81d1f8fb5cecd54e4ec3104763ba937aa7692a86dc5",
		1722479,
		"00000000000000ca7a4633264b9989355e9709f9e9da19506b0f636cc435dc8f")
	c.Assert(blockMetaAccessor.SaveBlockMeta(blockMeta.Height, blockMeta), IsNil)

	key := blockMetaAccessor.getBlockMetaKey(blockMeta.Height)
	c.Assert(key, Equals, fmt.Sprintf(PrefixBlocMeta+"%d", blockMeta.Height))

	bm, err := blockMetaAccessor.GetBlockMeta(blockMeta.Height)
	c.Assert(err, IsNil)
	c.Assert(bm, NotNil)

	nbm, err := blockMetaAccessor.GetBlockMeta(1024)
	c.Assert(err, IsNil)
	c.Assert(nbm, IsNil)

	for i := 0; i < 1024; i++ {
		bm := NewBlockMeta(thorchain.GetRandomTxHash().String(), int64(i), thorchain.GetRandomTxHash().String())
		c.Assert(blockMetaAccessor.SaveBlockMeta(bm.Height, bm), IsNil)
	}
	blockMetas, err := blockMetaAccessor.GetBlockMetas()
	c.Assert(err, IsNil)
	c.Assert(blockMetas, HasLen, 1025)
	c.Assert(blockMetaAccessor.PruneBlockMeta(1000), IsNil)
	allBlockMetas, err := blockMetaAccessor.GetBlockMetas()
	c.Assert(err, IsNil)
	c.Assert(allBlockMetas, HasLen, 25)

	fee, vSize, err := blockMetaAccessor.GetTransactionFee()
	c.Assert(err, NotNil)
	c.Assert(fee, Equals, 0.0)
	c.Assert(vSize, Equals, int32(0))
	// upsert transaction fee
	c.Assert(blockMetaAccessor.UpsertTransactionFee(1.0, 1), IsNil)
	fee, vSize, err = blockMetaAccessor.GetTransactionFee()
	c.Assert(err, IsNil)
	c.Assert(fee, Equals, 1.0)
	c.Assert(vSize, Equals, int32(1))
}
