package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type AdminConfigSuite struct{}

var _ = Suite(&AdminConfigSuite{})

func (s AdminConfigSuite) TestGetKey(c *C) {
	keys := []string{"GSL", "StakerAmtInterval", "PoolAddress", "Unknown", "PoolExpiry", "MinStakerCoins", "MRRA", "MinValidatorBond", "WhiteListGasToken"}
	for _, key := range keys {
		c.Check(GetAdminConfigKey(key).String(), Equals, key)
	}
	c.Check(GetAdminConfigKey("bogus key"), Equals, UnknownKey)
}

func (s AdminConfigSuite) TestAdminConfig(c *C) {
	amts := []string{"GSL", "StakerAmtInterval"}
	addr, err := sdk.AccAddressFromBech32("bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Assert(err, IsNil)
	for _, amt := range amts {
		config := NewAdminConfig(GetAdminConfigKey(amt), "12", addr) // happy path
		c.Check(config.Valid(), IsNil, Commentf("%s", amt))
		config = NewAdminConfig(GetAdminConfigKey(amt), "abc", addr) // invalid value
		c.Check(config.Valid(), NotNil, Commentf("%s", amt))
	}
	uintAmnt := []string{"MRRA", "MinValidatorBond"}
	for _, item := range uintAmnt {
		cfg := NewAdminConfig(GetAdminConfigKey(item), "1000", addr)
		c.Check(cfg.Valid(), IsNil, Commentf("%s", item))
		cfg1 := NewAdminConfig(GetAdminConfigKey(item), "whatever", addr)
		c.Check(cfg1.Valid(), NotNil, Commentf("%s", item))
	}
	coinAmt := []string{"MinStakerCoins", "WhiteListGasToken"}
	for _, item := range coinAmt {
		cfg := NewAdminConfig(GetAdminConfigKey(item), "100bep", addr)
		c.Check(cfg.Valid(), IsNil, Commentf("%s is invalid coin", item))
		cfg1 := NewAdminConfig(GetAdminConfigKey(item), "1233", addr)
		c.Check(cfg1.Valid(), NotNil, Commentf("%s is not valid coin", item))
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
	c.Check(adminCfg.DbKey(), Equals, "GSL_bep1jtpv39zy5643vywg7a9w73ckg880lpwuqd444v")
	c.Check(adminCfg.String(), Equals, fmt.Sprintf("Config: %s --> %s", adminCfg.Key, adminCfg.Value))

	inputs := []struct {
		address sdk.AccAddress
		key     AdminConfigKey
		value   string
	}{
		{
			address: sdk.AccAddress{},
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
		{
			address: addr,
			key:     MinValidatorBondKey,
			value:   "blab",
		},
	}
	for _, item := range inputs {
		adminCfg := NewAdminConfig(item.key, item.value, item.address)
		c.Assert(adminCfg.Valid(), NotNil)
	}
}
