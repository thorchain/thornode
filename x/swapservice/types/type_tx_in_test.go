package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	common "gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type TypeTxInSuite struct{}

var _ = Suite(&TypeTxInSuite{})

func (s TypeTxInSuite) TestVoter(c *C) {
	txID, err := common.NewTxID("A1C7D97D5DB51FFDBC3FE29FFF6ADAA2DAF112D2CEAADA0902822333A59BD218")
	c.Assert(err, IsNil)
	txID2, err := common.NewTxID("47B4FE474A63DDF79DF2790C1C5162F4C213484750AB8292CFE7342E4B0B40E2")
	c.Assert(err, IsNil)
	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	acc1, err := sdk.AccAddressFromBech32("cosmos1jye8q836gf2ffw7tpc7kvp2uyaq56njfn4fpc4")
	c.Assert(err, IsNil)
	acc2, err := sdk.AccAddressFromBech32("cosmos1nm0rrq86ucezaf8uj35pq9fpwr5r82cl8sc7p5")
	c.Assert(err, IsNil)
	acc3, err := sdk.AccAddressFromBech32("cosmos13vf50slxyapl0a9ty9srz54g0pfmd7fr7pq26z")
	c.Assert(err, IsNil)

	voter := NewTxInVoter(txID, nil)

	txIn := NewTxIn(nil, "hello", bnb)
	txIn2 := NewTxIn(nil, "goodbye", bnb)

	voter.Adds([]TxIn{txIn}, acc1)
	c.Assert(voter.Txs, HasLen, 1)

	voter.Adds([]TxIn{txIn}, acc1) // check we don't duplicate the same signer
	c.Assert(voter.Txs, HasLen, 1)
	c.Assert(voter.Txs[0].Signers, HasLen, 1)

	voter.Add(txIn, acc2) // append a signature
	c.Assert(voter.Txs, HasLen, 1)
	c.Assert(voter.Txs[0].Signers, HasLen, 2)

	voter.Add(txIn2, acc1) // same validator seeing a different version of tx
	c.Assert(voter.Txs, HasLen, 1)
	c.Assert(voter.Txs[0].Signers, HasLen, 2)

	voter.Add(txIn2, acc3) // second version
	c.Assert(voter.Txs, HasLen, 2)
	c.Assert(voter.Txs[0].Signers, HasLen, 2)
	c.Assert(voter.Txs[1].Signers, HasLen, 1)

	tx := voter.GetTx(3)
	c.Check(tx.Memo, Equals, "hello")
	tx = voter.GetTx(4)
	c.Check(tx.Empty(), Equals, true)
	c.Check(voter.HasConensus(3), Equals, true)
	c.Check(voter.HasConensus(4), Equals, false)

	voter.SetDone(txID2)
	for _, transaction := range voter.Txs {
		c.Check(transaction.Done.Equals(txID2), Equals, true)
	}
}
