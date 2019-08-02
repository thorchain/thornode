package exchange

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

type TxHashSuite struct{}

var _ = Suite(&TxHashSuite{})

func (s *TxHashSuite) TestTxHash(c *C) {
	result, err := GetTxIfno("ED92EB231E176EF54CCF6C34E83E44BA971192E75D55C86953BF0FB371F042FA")
	c.Assert(err, IsNil)
	c.Check(result.Memo(), Equals, "test")
	c.Check(result.Inputs(), HasLen, 1)
	c.Check(result.Inputs()[0].Address, Equals, "tbnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7whxk9nt")
	c.Assert(result.Inputs()[0].Coins, HasLen, 1)
	c.Check(result.Inputs()[0].Coins[0].Amount.String(), Equals, "100000000")
	c.Check(result.Inputs()[0].Coins[0].Denom, Equals, "LOK-3C0")
	c.Check(result.Outputs(), HasLen, 1)
	c.Check(result.Outputs()[0].Address, Equals, "tbnb13wkwssdkxxj9ypwpgmkaahyvfw5qk823v8kqhl")
	c.Assert(result.Outputs()[0].Coins, HasLen, 1)
	c.Check(result.Outputs()[0].Coins[0].Amount.String(), Equals, "100000000")
	c.Check(result.Outputs()[0].Coins[0].Denom, Equals, "LOK-3C0")
	//fmt.Printf("%v\n", result.Inputs())
}
