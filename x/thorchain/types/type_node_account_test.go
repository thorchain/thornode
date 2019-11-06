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

func (s *NodeAccountSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

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
	bnb := GetRandomBNBAddress()
	addr := GetRandomBech32Addr()
	c.Check(addr.Empty(), Equals, false)
	bepConsPubKey := GetRandomBech32ConsensusPubKey()
	trustAccount := NewTrustAccount(bnb, addr, bepConsPubKey)
	err := trustAccount.IsValid()
	c.Assert(err, IsNil)
	pubkey, _ := common.NewPubKeyFromBech32(addr.String())
	bondAddr := GetRandomBNBAddress()
	na := NewNodeAccount(pubkey, Active, trustAccount, sdk.NewUint(common.One), bondAddr, 1)
	c.Assert(na.IsEmpty(), Equals, false)
	c.Assert(na.IsValid(), IsNil)
	c.Assert(na.PubKey.GetThorAddress().String(), Equals, na.GetNodeAddress().String())
	c.Assert(na.Bond.Uint64(), Equals, uint64(common.One))
	nas := NodeAccounts{
		na,
	}
	c.Assert(nas.IsTrustAccount(addr), Equals, true)
	naEmpty := NewNodeAccount(common.EmptyPubKey, Active, trustAccount, sdk.NewUint(common.One), bondAddr, 1)
	c.Assert(naEmpty.IsValid(), NotNil)
	c.Assert(naEmpty.IsEmpty(), Equals, true)
	invalidBondAddr := NewNodeAccount(common.EmptyPubKey, Active, trustAccount, sdk.NewUint(common.One), "", 1)
	c.Assert(invalidBondAddr.IsValid(), NotNil)
}

func (NodeAccountSuite) TestNodeAccountsSort(c *C) {
	var accounts NodeAccounts
	for {
		na := GetRandomNodeAccount(Active)
		dup := false
		for _, node := range accounts {
			if na.Accounts.SignerBNBAddress.Equals(node.Accounts.SignerBNBAddress) {
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
		if na.Accounts.SignerBNBAddress.String() < accounts[i].Accounts.SignerBNBAddress.String() {
			c.Errorf("%s should be before %s", na.Accounts.SignerBNBAddress, accounts[i].Accounts.SignerBNBAddress)
		}

	}
}

func (NodeAccountSuite) TestAfter(c *C) {
	var accounts NodeAccounts
	for {
		na := GetRandomNodeAccount(Active)
		dup := false
		for _, node := range accounts {
			if na.Accounts.SignerBNBAddress.Equals(node.Accounts.SignerBNBAddress) {
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
	for i := 0; i < len(accounts)-1; i++ {
		node := accounts[i]
		nextNode := accounts.After(node.Accounts.SignerBNBAddress)
		c.Assert(accounts[i+1].Accounts.SignerBNBAddress.String(), Equals, nextNode.Accounts.SignerBNBAddress.String())
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
