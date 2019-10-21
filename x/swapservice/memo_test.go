package swapservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type MemoSuite struct{}

var _ = Suite(&MemoSuite{})

func (s *MemoSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (s *MemoSuite) TestTxType(c *C) {
	for _, trans := range []TxType{txCreate, txStake, txWithdraw, txSwap, txOutbound, txAdd, txGas, txBond, txLeave, txNextPool, txAck} {
		tx, err := stringToTxType(trans.String())
		c.Assert(err, IsNil)
		c.Check(tx, Equals, trans)
	}
}
func (s *MemoSuite) TestParseWithAbbreviated(c *C) {
	// happy paths
	memo, err := ParseMemo("c:RUNE-1BA")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txCreate), Equals, true, Commentf("MEMO: %+v", memo))

	memo, err = ParseMemo("%:RUNE-1BA")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txAdd), Equals, true, Commentf("MEMO: %+v", memo))

	memo, err = ParseMemo("+:RUNE-1BA")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txStake), Equals, true, Commentf("MEMO: %+v", memo))

	memo, err = ParseMemo("-:RUNE-1BA:25")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txWithdraw), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetAmount(), Equals, "25")

	memo, err = ParseMemo("=:RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:870000000")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Log(memo.GetSlipLimit().Uint64())
	c.Check(memo.GetSlipLimit().Equal(sdk.NewUint(870000000)), Equals, true)

	memo, err = ParseMemo("=:RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(memo.GetSlipLimit().Uint64(), Equals, uint64(0))

	memo, err = ParseMemo("=:RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(memo.GetSlipLimit().Equal(sdk.ZeroUint()), Equals, true)

	// unhappy paths
	_, err = ParseMemo("")
	c.Assert(err, NotNil)
	_, err = ParseMemo("bogus")
	c.Assert(err, NotNil)
	_, err = ParseMemo("CREATE") // missing symbol
	c.Assert(err, NotNil)
	_, err = ParseMemo("c:") // bad symbol
	c.Assert(err, NotNil)
	_, err = ParseMemo("-:bnb") // withdraw basis points is optional
	c.Assert(err, IsNil)
	_, err = ParseMemo("-:bnb:twenty-two") // bad amount
	c.Assert(err, NotNil)
	_, err = ParseMemo("=:bnb:bad_DES:5.6") // bad destination
	c.Assert(err, NotNil)
	_, err = ParseMemo(">:bnb:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:five") // bad slip limit
	c.Assert(err, NotNil)
	_, err = ParseMemo("!:key:val") // not enough arguments
	c.Assert(err, NotNil)
	_, err = ParseMemo("!:bogus:key:value") // bogus admin command type
	c.Assert(err, NotNil)
	_, err = ParseMemo("leave:whatever")
	c.Assert(err, NotNil)
	_, err = ParseMemo("ack:whatever")
	c.Assert(err, NotNil)
	_, err = ParseMemo("nextpool:whatever")
	c.Assert(err, NotNil)
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

	memo, err = ParseMemo("SWAP:RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:870000000")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Log(memo.GetSlipLimit().String())
	c.Check(memo.GetSlipLimit().Equal(sdk.NewUint(870000000)), Equals, true)

	memo, err = ParseMemo("SWAP:RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(memo.GetSlipLimit().Uint64(), Equals, uint64(0))

	memo, err = ParseMemo("SWAP:RUNE-1BA:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6:")
	c.Assert(err, IsNil)
	c.Check(memo.GetTicker().String(), Equals, "RUNE-1BA")
	c.Check(memo.IsType(txSwap), Equals, true, Commentf("MEMO: %+v", memo))
	c.Check(memo.GetDestination().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Check(memo.GetSlipLimit().Uint64(), Equals, uint64(0))

	whiteListAddr := GetRandomBech32Addr()
	memo, err = ParseMemo("bond:" + whiteListAddr.String())
	c.Assert(err, IsNil)
	c.Assert(memo.IsType(txBond), Equals, true)
	c.Assert(memo.GetNodeAddress().String(), Equals, whiteListAddr.String())

	memo, err = ParseMemo("nextpool:bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	c.Assert(err, IsNil)
	c.Assert(memo.IsType(txNextPool), Equals, true)
	c.Assert(memo.GetNextPoolAddress().String(), Equals, "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")

	memo, err = ParseMemo("leave")
	c.Assert(err, IsNil)
	c.Assert(memo.IsType(txLeave), Equals, true)
	memo, err = ParseMemo("ack")
	c.Assert(err, IsNil)
	c.Assert(memo.IsType(txAck), Equals, true)

	// unhappy paths
	_, err = ParseMemo("")
	c.Assert(err, NotNil)
	_, err = ParseMemo("bogus")
	c.Assert(err, NotNil)
	_, err = ParseMemo("CREATE") // missing symbol
	c.Assert(err, NotNil)
	_, err = ParseMemo("CREATE:") // bad symbol
	c.Assert(err, NotNil)
	_, err = ParseMemo("withdraw:bnb") // withdraw basis points is optional
	c.Assert(err, IsNil)
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
