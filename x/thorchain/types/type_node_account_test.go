package types

import (
	"encoding/json"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"
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
		"disabled":    Disabled,
		"Disabled":    Disabled,
		"disabLed":    Disabled,
		"ready":       Ready,
		"Ready":       Ready,
		"rEady":       Ready,
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
	addr := GetRandomBech32Addr()
	c.Check(addr.Empty(), Equals, false)
	bepConsPubKey := GetRandomBech32ConsensusPubKey()
	nodeAddress := GetRandomBech32Addr()
	bondAddr := GetRandomBNBAddress()
	pubKeys := common.PubKeys{
		Secp256k1: GetRandomPubKey(),
		Ed25519:   GetRandomPubKey(),
	}

	na := NewNodeAccount(nodeAddress, Active, pubKeys, bepConsPubKey, sdk.NewUint(common.One), bondAddr, 1)
	c.Assert(na.IsEmpty(), Equals, false)
	c.Assert(na.IsValid(), IsNil)
	c.Assert(na.Bond.Uint64(), Equals, uint64(common.One))
	nas := NodeAccounts{
		na,
	}
	c.Assert(nas.IsTrustAccount(addr), Equals, false)
	c.Assert(nas.IsTrustAccount(nodeAddress), Equals, true)
	c.Logf("node account:%s", na)
	naEmpty := NewNodeAccount(sdk.AccAddress{}, Active, pubKeys, bepConsPubKey, sdk.NewUint(common.One), bondAddr, 1)
	c.Assert(naEmpty.IsValid(), NotNil)
	c.Assert(naEmpty.IsEmpty(), Equals, true)
	invalidBondAddr := NewNodeAccount(sdk.AccAddress{}, Active, pubKeys, bepConsPubKey, sdk.NewUint(common.One), "", 1)
	c.Assert(invalidBondAddr.IsValid(), NotNil)
}

func (NodeAccountSuite) TestNodeAccountsSort(c *C) {
	var accounts NodeAccounts
	for {
		na := GetRandomNodeAccount(Active)
		dup := false
		for _, node := range accounts {
			if na.NodeAddress.Equals(node.NodeAddress) {
				dup = true
			}
		}
		if dup {
			continue
		}
		accounts = append(accounts, na)
		if len(accounts) == 10 {
			break
		}
	}

	sort.Sort(accounts)

	for i, na := range accounts {
		if i == 0 {
			continue
		}
		if na.NodeAddress.String() < accounts[i].NodeAddress.String() {
			c.Errorf("%s should be before %s", na.NodeAddress, accounts[i].NodeAddress)
		}

	}
}

func (NodeAccountSuite) TestNodeAccountUpdateStatusAndSort(c *C) {
	var accounts NodeAccounts
	for i := 0; i < 10; i++ {
		na := GetRandomNodeAccount(Active)
		accounts = append(accounts, na)
	}
	isSorted := sort.SliceIsSorted(accounts, func(i, j int) bool {
		return accounts[i].StatusSince < accounts[j].StatusSince
	})
	c.Assert(isSorted, Equals, true)
}
