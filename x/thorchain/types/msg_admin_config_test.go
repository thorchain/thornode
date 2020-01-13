package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type MsgSetAdminConfigSuite struct{}

var _ = Suite(&MsgSetAdminConfigSuite{})

func (MsgSetAdminConfigSuite) TestMsgSetAdminConfig(c *C) {
	addr := GetRandomBech32Addr()
	c.Check(addr.Empty(), Equals, false)
	tx := GetRandomTx()
	msgSetAdminConfig := NewMsgSetAdminConfig(tx, DefaultPoolStatus, "Enabled", addr)
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
		value  string
	}{
		{
			signer: sdk.AccAddress{},
			value:  "1.0",
		},
		{
			signer: addr,
			value:  "helloWorld",
		},
	}

	for _, item := range inputs {
		m := NewMsgSetAdminConfig(tx, DefaultPoolStatus, item.value, item.signer)
		err := m.ValidateBasic()
		c.Assert(err, NotNil)
	}
}
