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

func (s *MemoSuite) TestValidateDestination(c *C) {
	c.Check(validateDestination("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6"), IsNil)
	c.Check(validateDestination("tbnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6"), IsNil)
	c.Check(validateDestination("1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6"), NotNil)
	c.Check(validateDestination("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6X"), NotNil)
	c.Check(validateDestination("bogus"), NotNil)
}

func (s *MemoSuite) TestValidateSymbol(c *C) {
	c.Check(validateSymbol("BNB"), IsNil)
	c.Check(validateSymbol("RUNE-1BA"), IsNil)
	c.Check(validateSymbol(""), NotNil)
	c.Check(validateSymbol("this is way tooo long"), NotNil)
}

func (s *MemoSuite) TestParse(c *C) {
	// happy paths
	memo, err := ParseMemo("CREATE:RUNE-1BA")
	c.Assert(err, IsNil)
	c.Check(memo.GetSymbol(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txCreate), Equals, true, Commentf("MEMO: %+v", memo))

	memo, err = ParseMemo("STAKE:RUNE-1BA")
	c.Assert(err, IsNil)
	c.Check(memo.GetSymbol(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txStake), Equals, true, Commentf("MEMO: %+v", memo))

	memo, err = ParseMemo("WITHDRAW:RUNE-1BA:25")
	c.Assert(err, IsNil)
	c.Check(memo.GetSymbol(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txWithdraw), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetAmount(), Equals, 25.0)

	memo, err = ParseMemo("SWAP:RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:8.7:hello to : : the world!")
	c.Assert(err, IsNil)
	c.Check(memo.GetSymbol(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(memo.GetSlipLimit(), Equals, 8.7)
	c.Check(memo.GetMemo(), Equals, "hello to : : the world!")

	memo, err = ParseMemo("SWAP:RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Assert(err, IsNil)
	c.Check(memo.GetSymbol(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(memo.GetSlipLimit(), Equals, 0.0)
	c.Check(memo.GetMemo(), Equals, "")

	memo, err = ParseMemo("SWAP:RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6::hi")
	c.Assert(err, IsNil)
	c.Check(memo.GetSymbol(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(memo.GetSlipLimit(), Equals, 0.0)
	c.Check(memo.GetMemo(), Equals, "hi")

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
}
