package types

import (
	"encoding/json"

	. "gopkg.in/check.v1"
)

type TxOutTestSuite struct{}

var _ = Suite(&TxOutTestSuite{})

func (TxOutTestSuite) TestTxOut(c *C) {
	input := `{ "height": "1718", "hash": "", "tx_array": [ { "chain": "BNB", "in_hash": "9999A5A08D8FCF942E1AAAA01AB1E521B699BA3A009FA0591C011DC1FFDC5E68", "to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "memo": "REFUND:TODO", "coin":  { "asset": "BNB.BNB", "amount": "194765912" }  } ]}`
	var item TxOut
	err := json.Unmarshal([]byte(input), &item)
	c.Check(err, IsNil)
	c.Check(len(item.TxArray), Equals, 1)
	c.Check(item.TxArray[0].Coin.Amount.Uint64(), Equals, uint64(194765912))
	c.Check(item.TxArray[0].Coin.Asset.IsBNB(), Equals, true)
	c.Check(item.TxArray[0].TxOutItem().Hash(), Equals, "53F21B135F9E520DC442BA4EEF1870AB77D40F4EAD77BE72E35CBE06697D46E2")

	input = `{ "height": "1718", "hash": "", "tx_array": [ { "chain": "BNB", "in_hash": "9999A5A08D8FCF942E1AAAA01AB1E521B699BA3A009FA0591C011DC1FFDC5E68", "to": "tbnb1yxfyeda8pnlxlmx0z3cwx74w9xevspwdpzdxpj", "memo": "REFUND:TODO" } ]}`
	var item2 TxOut
	err = json.Unmarshal([]byte(input), &item2)
	c.Check(err, IsNil)
	c.Check(len(item2.TxArray), Equals, 1)
	c.Check(item2.TxArray[0].TxOutItem().Hash(), Equals, "680F2FA2EEB80C730C5723799E8F82F316E9C6315FFC22C32212A2B18B22B6A9")
}
