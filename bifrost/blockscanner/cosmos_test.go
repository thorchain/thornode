package blockscanner

import (
	"encoding/json"

	. "gopkg.in/check.v1"
)

type CosmosSupplementalSuite struct{}

var _ = Suite(&CosmosSupplementalSuite{})

func (s *CosmosSupplementalSuite) TestBlockRequest(c *C) {
	supp := CosmosSupplemental{}
	url, body := supp.BlockRequest("http://localhost:222/block", 23)
	c.Check(body, Equals, "")
	c.Check(url, Equals, "http://localhost:222/block?height=23")
}

func (s *CosmosSupplementalSuite) TestUnmarshalBlock(c *C) {
	block := item{
		Result: itemResult{
			Block: itemBlock{
				Header: itemHeader{
					Height: "400",
				},
				Data: itemData{
					Txs: []string{"a", "b", "c"},
				},
			},
		},
	}

	bz, err := json.Marshal(block)
	c.Assert(err, IsNil)

	supp := CosmosSupplemental{}
	height, txns, err := supp.UnmarshalBlock(bz)
	c.Assert(err, IsNil)
	c.Check(height, Equals, int64(400))
	c.Check(txns, DeepEquals, []string{"a", "b", "c"})

	height, txns, err = supp.UnmarshalBlock([]byte(`{ "jsonrpc": "2.0", "id": "", "result": { "block_meta": null, "block": null } }`))
	c.Assert(err, IsNil)
	c.Check(height, Equals, int64(0))
	c.Check(txns, HasLen, 0)
}
