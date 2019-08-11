package types

import (
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
	for _, amt := range amts {
		config := NewAdminConfig(GetAdminConfigKey(amt), "12") // happy path
		c.Check(config.Valid(), IsNil, Commentf("%s", amt))
		config = NewAdminConfig(GetAdminConfigKey(amt), "abc") // invalid value
		c.Check(config.Valid(), NotNil, Commentf("%s", amt))
	}
	bnbs := []string{"PoolAddress"}
	for _, bnb := range bnbs {
		config := NewAdminConfig(GetAdminConfigKey(bnb), "bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6") // happy path
		c.Check(config.Valid(), IsNil, Commentf("%s", bnb))
		config = NewAdminConfig(GetAdminConfigKey(bnb), "abc") // invalid value
		c.Check(config.Valid(), NotNil, Commentf("%s", bnb))
	}
}
