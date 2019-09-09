package types

import (
	"fmt"

	"gitlab.com/thorchain/bepswap/common"
	. "gopkg.in/check.v1"
)

type AdminConfigSuite struct{}

var _ = Suite(&AdminConfigSuite{})

func (s AdminConfigSuite) TestGetKey(c *C) {
	keys := []string{"GSL", "TSL", "StakerAmtInterval", "PoolAddress", "Unknown"}
	for _, key := range keys {
		c.Check(GetAdminConfigKey(key).String(), Equals, key)
	}
	c.Check(GetAdminConfigKey("bogus key"), Equals, UnknownKey)
}

func (s AdminConfigSuite) TestAdminConfig(c *C) {
	amts := []string{"GSL", "TSL", "StakerAmtInterval"}
	addr, err := common.NewBnbAddress("bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Assert(err, IsNil)
	for _, amt := range amts {
		config := NewAdminConfig(GetAdminConfigKey(amt), "12", addr) // happy path
		c.Check(config.Valid(), IsNil, Commentf("%s", amt))
		config = NewAdminConfig(GetAdminConfigKey(amt), "abc", addr) // invalid value
		c.Check(config.Valid(), NotNil, Commentf("%s", amt))
	}
	bnbs := []string{"PoolAddress"}
	for _, bnb := range bnbs {
		config := NewAdminConfig(GetAdminConfigKey(bnb), "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6", addr) // happy path
		c.Check(config.Valid(), IsNil, Commentf("%s", bnb))
		config = NewAdminConfig(GetAdminConfigKey(bnb), "abc", addr) // invalid value
		c.Check(config.Valid(), NotNil, Commentf("%s", bnb))
	}
	adminCfg := NewAdminConfig(GSLKey, "100", addr)
	c.Check(adminCfg.Empty(), Equals, false)
	c.Check(adminCfg.DbKey(), Equals, "GSL_bnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
	c.Check(adminCfg.String(), Equals, fmt.Sprintf("Config: %s --> %s", adminCfg.Key, adminCfg.Value))

	inputs := []struct {
		address common.BnbAddress
		key     AdminConfigKey
		value   string
	}{
		{
			address: common.NoBnbAddress,
			key:     GSLKey,
			value:   "1.0",
		},
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
			key:     GSLKey,
			value:   "",
		},
		{
			address: addr,
			key:     GSLKey,
			value:   "nothing",
		},
		{
			address: addr,
			key:     PoolAddressKey,
			value:   "hahaha",
		},
	}
	for _, item := range inputs {
		adminCfg := NewAdminConfig(item.key, item.value, item.address)
		c.Assert(adminCfg.Valid(), NotNil)
	}
}
