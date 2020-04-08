package types

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type EthereumSupplementalSuite struct{}

var _ = Suite(&EthereumSupplementalSuite{})

func (s *EthereumSupplementalSuite) TestBlockRequest(c *C) {
	supp := EthereumSupplemental{}
	url, body := supp.BlockRequest("http://localhost:222", 23)
	c.Check(body, Equals, `{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x17", true],"id":1}`)
	c.Check(url, Equals, "http://localhost:222")
}

func (s *EthereumSupplementalSuite) TestUnmarshalBlock(c *C) {
	blockJson := `{
		"difficulty": "0x31962a3fc82b",
		"extraData": "0x4477617266506f6f6c",
		"gasLimit": "0x47c3d8",
		"gasUsed": "0x0",
		"hash": "0x78bfef68fccd4507f9f4804ba5c65eb2f928ea45b3383ade88aaa720f1209cba",
		"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		"miner": "0x2a65aca4d5fc5b5c859090a6c34d164135398226",
		"nonce": "0xa5e8fb780cc2cd5e",
		"number": "0x17",
		"parentHash": "0x8b535592eb3192017a527bbf8e3596da86b3abea51d6257898b2ced9d3a83826",
		"receiptsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
		"sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
		"size": "0x20e",
		"stateRoot": "0xdc6ed0a382e50edfedb6bd296892690eb97eb3fc88fd55088d5ea753c48253dc",
		"timestamp": "0x579f4981",
		"totalDifficulty": "0x25cff06a0d96f4bee",
		"transactions": [{
			"blockHash":"0x78bfef68fccd4507f9f4804ba5c65eb2f928ea45b3383ade88aaa720f1209cba",
			"blockNumber":"0x17",
			"from":"0xa7d9ddbe1f17865597fbd27ec712455208b6b76d",
			"gas":"0xc350",
			"gasPrice":"0x4a817c800",
			"hash":"0x88df016429689c079f3b2f6ad39fa052532c56795b733da78a91ebe6a713944b",
			"input":"0x68656c6c6f21",
			"nonce":"0x15",
			"to":"0xf02c1c8e6114b1dbe8937a39260b5b0a374432bb",
			"transactionIndex":"0x41",
			"value":"0xf3dbb76162000",
			"v":"0x25",
			"r":"0x1b5e176d927f8e9ab405058b2d2457392da3e20f328b16ddabcebc33eaac5fea",
			"s":"0x4ba69724e8f69de52f0125ad8b3c5c2cef33019bac3249e2c0a2192766d1721c"
		}],
		"transactionsRoot": "0x88df016429689c079f3b2f6ad39fa052532c56795b733da78a91ebe6a713944b",
		"uncles": []
	}`

	supp := EthereumSupplemental{}
	txns, err := supp.UnmarshalBlock([]byte(blockJson))
	c.Assert(err, IsNil)
	c.Check(txns, DeepEquals, []string{`{"nonce":"0x15","gasPrice":"0x4a817c800","gas":"0xc350","to":"0xf02c1c8e6114b1dbe8937a39260b5b0a374432bb","value":"0xf3dbb76162000","input":"0x68656c6c6f21","v":"0x25","r":"0x1b5e176d927f8e9ab405058b2d2457392da3e20f328b16ddabcebc33eaac5fea","s":"0x4ba69724e8f69de52f0125ad8b3c5c2cef33019bac3249e2c0a2192766d1721c","hash":"0x88df016429689c079f3b2f6ad39fa052532c56795b733da78a91ebe6a713944b"}`})
}
