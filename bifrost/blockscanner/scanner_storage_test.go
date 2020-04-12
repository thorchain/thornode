package blockscanner

import (
	. "gopkg.in/check.v1"
)

type BlockScannerStorageSuite struct{}

var _ = Suite(&BlockScannerStorageSuite{})

func (s *BlockScannerStorageSuite) TestScannerSetup(c *C) {
	tmpdir := "/tmp/scanner_storage"
	scanner, err := NewBlockScannerStorage(tmpdir)
	c.Assert(err, IsNil)
	c.Assert(scanner, NotNil)

	// in memory storage
	scanner, err = NewBlockScannerStorage("")
	c.Assert(err, IsNil)
	c.Assert(scanner, NotNil)
}
