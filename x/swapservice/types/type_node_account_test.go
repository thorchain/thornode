package types

import (
	"encoding/json"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type NodeAccountSuite struct{}

var _ = Suite(&NodeAccountSuite{})

func (NodeAccountSuite) TestGetNodeStatus(c *C) {
	input := map[string]NodeStatus{
		"unknown":     Unknown,
		"Unknown":     Unknown,
		"uNknown":     Unknown,
		"WhiteListed": WhiteListed,
		"WHITELISTED": WhiteListed,
		"whitelisted": WhiteListed,
		"Standby":     Standby,
		"standby":     Standby,
		"StanDby":     Standby,
		"active":      Active,
		"Active":      Active,
		"aCtive":      Active,
		"ACTIVE":      Active,
		"queued":      Queued,
		"Queued":      Queued,
		"qUeued":      Queued,
		"disabled":    Disabled,
		"Disabled":    Disabled,
		"disabLed":    Disabled,
	}
	for k, v := range input {
		r := GetNodeStatus(k)
		if r != v {
			c.Errorf("expect %s,however we got %s", v, r)
		}
		c.Check(r.String(), Equals, strings.ToLower(k))
		c.Check(v.Valid(), IsNil)
		buf, err := json.Marshal(v)
		c.Assert(err, IsNil)
		c.Assert(len(buf) > 0, Equals, true)
	}
	ns := NodeStatus(255)
	c.Assert(ns.String(), Equals, "")
	c.Check(ns.Valid(), NotNil)
	ns = GetNodeStatus("Whatever")
	c.Assert(ns, Equals, Unknown)
}

func (NodeAccountSuite) TestNodeAccount(c *C) {
	bnb, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	addr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	c.Check(addr.Empty(), Equals, false)
	bepConsPubKey := `bepcpub1zcjduepq4kn64fcjhf0fp20gp8var0rm25ca9jy6jz7acem8gckh0nkplznq85gdrg`
	trustAccount := NewTrustAccount(bnb, addr, bepConsPubKey)
	err = trustAccount.IsValid()
	c.Assert(err, IsNil)
	nodeAddress, err := sdk.AccAddressFromBech32("bep1rtgz3lcaw8vw0yfsc8ga0rdgwa3qh9ju7vfsnk")
	c.Assert(err, IsNil)
	na := NewNodeAccount(nodeAddress, Active, trustAccount)
	na.Bond = sdk.NewUint(common.One)
	c.Assert(na.IsEmpty(), Equals, false)
	c.Assert(na.IsValid(), IsNil)
	c.Assert(na.Bond.Uint64(), Equals, uint64(common.One))
	nas := NodeAccounts{
		na,
	}
	c.Assert(nas.IsTrustAccount(addr), Equals, true)
	c.Assert(nas.IsTrustAccount(nodeAddress), Equals, false)
	c.Logf("node account:%s", na)
	naEmpty := NewNodeAccount(sdk.AccAddress{}, Active, trustAccount)
	c.Assert(naEmpty.IsValid(), NotNil)
	c.Assert(naEmpty.IsEmpty(), Equals, true)

}
