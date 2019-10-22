package types

import (
	"encoding/json"

	. "gopkg.in/check.v1"
)

type TxOutTestSuite struct{}

var _ = Suite(&TxOutTestSuite{})

func (TxOutTestSuite) TestTxOut(c *C) {
	input := `{ "height": "1718", "hash": "", "tx_array": [ { "to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "coins": [ { "denom": "BNB", "amount": "194765912" } ] } ]}`
	var item TxOut
	err := json.Unmarshal([]byte(input), &item)
	c.Check(err, IsNil)
	c.Check(len(item.TxArray), Equals, 1)
	c.Check(item.TxArray[0].Coins[0].Amount.Uint64(), Equals, uint64(194765912))
}
