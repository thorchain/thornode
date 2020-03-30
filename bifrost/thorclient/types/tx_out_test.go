package types

import (
	"encoding/json"

	. "gopkg.in/check.v1"
)

type TxOutTestSuite struct{}

var _ = Suite(&TxOutTestSuite{})

func (TxOutTestSuite) TestTxOut(c *C) {
	input := `{ "height": "1718", "hash": "", "tx_array": [ { "chain": "BNB", "in_hash": "9999A5A08D8FCF942E1AAAA01AB1E521B699BA3A009FA0591C011DC1FFDC5E68", "to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "memo": "REFUND:TODO", "coin":  { "asset": "BNB", "amount": "194765912" }  } ]}`
	var item TxOut
	err := json.Unmarshal([]byte(input), &item)
	c.Check(err, IsNil)
	c.Check(len(item.TxArray), Equals, 1)
	c.Check(item.TxArray[0].Coin.Amount.Uint64(), Equals, uint64(194765912))
	c.Check(item.TxArray[0].Coin.Asset.IsBNB(), Equals, true)
	c.Check(item.TxArray[0].TxOutItem().Hash(), Equals, "DC247BAAC3376AC7102368AEC6CC6738C7813F6AC7B629DA8327CB9E10C667E1")

	input = `{ "height": "1718", "hash": "", "tx_array": [ { "chain": "BNB", "in_hash": "9999A5A08D8FCF942E1AAAA01AB1E521B699BA3A009FA0591C011DC1FFDC5E68", "to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "memo": "REFUND:TODO" } ]}`
	var item2 TxOut
	err = json.Unmarshal([]byte(input), &item2)
	c.Check(err, IsNil)
	c.Check(len(item2.TxArray), Equals, 1)
	c.Check(item2.TxArray[0].TxOutItem().Hash(), Equals, "6BCA5232893B143D50E1F108BED799789F8D09C853D4FA0DF8D54AE5F573DCC1")
}
