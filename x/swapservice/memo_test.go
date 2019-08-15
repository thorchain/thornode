package swapservice

import (
	. "gopkg.in/check.v1"
)

type MemoSuite struct{}

var _ = Suite(&MemoSuite{})

func (s *MemoSuite) TestTxType(c *C) {
	for _, trans := range []txType{txCreate, txStake, txWithdraw, txSwap} {
		tx, err := stringToTxType(trans.String())
		c.Assert(err, IsNil)
		c.Check(tx, Equals, trans)
	}
}

func (s *MemoSuite) TestParse(c *C) {
	// happy paths
	memo, err := ParseMemo("CREATE:RUNE-1BA")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txCreate), Equals, true, Commentf("MEMO: %+v", memo))

	memo, err = ParseMemo("add:RUNE-1BA")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txAdd), Equals, true, Commentf("MEMO: %+v", memo))

	memo, err = ParseMemo("STAKE:RUNE-1BA")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txStake), Equals, true, Commentf("MEMO: %+v", memo))

	memo, err = ParseMemo("WITHDRAW:RUNE-1BA:25")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txWithdraw), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetAmount(), Equals, "25")

	memo, err = ParseMemo("SWAP:RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:8.7:hello to : : the world!")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(memo.GetSlipLimit(), Equals, 8.7)
	c.Check(memo.GetMemo(), Equals, "hello to : : the world!")

	memo, err = ParseMemo("SWAP:RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(memo.GetSlipLimit(), Equals, 0.0)
	c.Check(memo.GetMemo(), Equals, "")

	memo, err = ParseMemo("SWAP:RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6::hi")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(memo.GetSlipLimit(), Equals, 0.0)
	c.Check(memo.GetMemo(), Equals, "hi")

	memo, err = ParseMemo("ADMIN:KEY:TSL:15")
	c.Assert(err, IsNil)
	c.Check(memo.GetAdminType(), Equals, adminKey)
	c.Check(memo.GetKey(), Equals, "TSL")
	c.Check(memo.GetValue(), Equals, "15")

	memo, err = ParseMemo("ADMIN:poolstatus:BNB:active")
	c.Assert(err, IsNil)
	c.Check(memo.GetAdminType(), Equals, adminPoolStatus)
	c.Check(memo.GetKey(), Equals, "BNB")
	c.Check(memo.GetValue(), Equals, "active")

	// unhappy paths
	_, err = ParseMemo("")
	c.Assert(err, NotNil)
	_, err = ParseMemo("bogus")
	c.Assert(err, NotNil)
	_, err = ParseMemo("CREATE") // missing symbol
	c.Assert(err, NotNil)
	_, err = ParseMemo("CREATE:") // bad symbol
	c.Assert(err, NotNil)
	_, err = ParseMemo("withdraw:bnb") // no amount
	c.Assert(err, NotNil)
	_, err = ParseMemo("withdraw:bnb:twenty-two") // bad amount
	c.Assert(err, NotNil)
	_, err = ParseMemo("swap:bnb:bad_DES:5.6") // bad destination
	c.Assert(err, NotNil)
	_, err = ParseMemo("swap:bnb:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:five") // bad slip limit
	c.Assert(err, NotNil)
	_, err = ParseMemo("admin:key:val") // not enough arguments
	c.Assert(err, NotNil)
	_, err = ParseMemo("admin:bogus:key:value") // bogus admin command type
	c.Assert(err, NotNil)
}
