package blockscanner

import (
	. "gopkg.in/check.v1"
)

type BlockScannerStorageSuite struct{}

var _ = Suite(&BlockScannerStorageSuite{})

func (s *BlockScannerStorageSuite) TestScannerSetup(c *C) {
	scanner, err := NewBlockScannerStorage("/tmp/scanner_storage")
	c.Assert(err, IsNil)
	c.Assert(scanner, NotNil)

	scanner, err = NewBlockScannerStorage("")
	c.Assert(err, NotNil)
	c.Assert(scanner, IsNil)
}
