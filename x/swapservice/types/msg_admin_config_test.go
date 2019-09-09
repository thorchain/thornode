package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type MsgSetAdminConfigSuite struct{}

var _ = Suite(&MsgSetAdminConfigSuite{})

func (MsgSetAdminConfigSuite) TestMsgSetAdminConfig(c *C) {
	bnbAddr, err := common.NewBnbAddress("tbnb1yycn4mh6ffwpjf584t8lpp7c27ghu03gpvqkfj")
	c.Assert(err, IsNil)
	c.Assert(bnbAddr.IsEmpty(), Equals, false)
	addr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	c.Check(addr.Empty(), Equals, false)
	msgSetAdminConfig := NewMsgSetAdminConfig(GSLKey, "2.0", bnbAddr, addr)
	c.Assert(msgSetAdminConfig.ValidateBasic(), IsNil)
	buf := msgSetAdminConfig.GetSignBytes()
	c.Assert(buf, NotNil)
	c.Check(len(buf) > 0, Equals, true)
	signer := msgSetAdminConfig.GetSigners()
	c.Assert(signer, NotNil)
	c.Check(len(signer) > 0, Equals, true)
	c.Check(msgSetAdminConfig.Route(), Equals, ModuleName)
	c.Check(msgSetAdminConfig.Type(), Equals, "set_admin_config")

	inputs := []struct {
		signer sdk.AccAddress
		from   common.BnbAddress
		value  string
	}{
		{
			signer: sdk.AccAddress{},
			from:   bnbAddr,
			value:  "1.0",
		},
		{
			signer: addr,
			from:   common.NoBnbAddress,
			value:  "1.0",
		},
		{
			signer: addr,
			from:   bnbAddr,
			value:  "helloWorld",
		},
	}

	for _, item := range inputs {
		m := NewMsgSetAdminConfig(GSLKey, item.value, item.from, item.signer)
		err := m.ValidateBasic()
		c.Assert(err, NotNil)
	}
}
