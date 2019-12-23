package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type AdminConfigSuite struct{}

var _ = Suite(&AdminConfigSuite{})

func (s AdminConfigSuite) TestGetKey(c *C) {
	keys := []string{
		"Unknown",
		"DefaultPoolStatus",
	}
	for _, key := range keys {
		c.Check(GetAdminConfigKey(key).String(), Equals, key)
	}
	c.Check(GetAdminConfigKey("bogus key"), Equals, UnknownKey)
}

func (s AdminConfigSuite) TestAdminConfig(c *C) {
	addr := GetRandomBech32Addr()

	inputs := []struct {
		address sdk.AccAddress
		key     AdminConfigKey
		value   string
	}{
		{
			address: addr,
			key:     "",
			value:   "1.0",
		},
		{
			address: addr,
			key:     UnknownKey,
			value:   "1.0",
		},

		{
			address: addr,
			key:     DefaultPoolStatus,
			value:   "123",
		},
	}
	for _, item := range inputs {
		adminCfg := NewAdminConfig(item.key, item.value, item.address)
		c.Assert(adminCfg.Valid(), NotNil)
	}
}

func (AdminConfigSuite) TestDefault(c *C) {
	input := map[AdminConfigKey]string{
		DefaultPoolStatus: "Enabled",
	}
	for k, v := range input {
		if k.Default() != v {
			c.Errorf("expected: %s , however THORNode got: %s", v, k.Default())
		}
	}
}
