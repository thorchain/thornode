package types

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

type AdminConfigSuite struct{}

var _ = Suite(&AdminConfigSuite{})

func (s AdminConfigSuite) TestGetKey(c *C) {
	keys := []string{
		"GSL",
		"StakerAmtInterval",
		"Unknown",
		"MinValidatorBond",
		"WhiteListGasAsset",
		"PoolRefundGas",
		"DefaultPoolStatus",
	}
	for _, key := range keys {
		c.Check(GetAdminConfigKey(key).String(), Equals, key)
	}
	c.Check(GetAdminConfigKey("bogus key"), Equals, UnknownKey)
}

func (s AdminConfigSuite) TestAdminConfig(c *C) {
	amts := []string{"GSL", "StakerAmtInterval"}
	addr := GetRandomBech32Addr()
	for _, amt := range amts {
		config := NewAdminConfig(GetAdminConfigKey(amt), "12", addr) // happy path
		c.Check(config.Valid(), IsNil, Commentf("%s", amt))
		config = NewAdminConfig(GetAdminConfigKey(amt), "abc", addr) // invalid value
		c.Check(config.Valid(), NotNil, Commentf("%s", amt))
	}
	uintAmnt := []string{
		"MinValidatorBond", "PoolRefundGas"}
	for _, item := range uintAmnt {
		cfg := NewAdminConfig(GetAdminConfigKey(item), "1000", addr)
		c.Check(cfg.Valid(), IsNil, Commentf("%s", item))
		cfg1 := NewAdminConfig(GetAdminConfigKey(item), "whatever", addr)
		c.Check(cfg1.Valid(), NotNil, Commentf("%s", item))
	}

	coinAmt := []string{"WhiteListGasAsset"}
	for _, item := range coinAmt {
		cfg := NewAdminConfig(GetAdminConfigKey(item), "100bep", addr)
		c.Check(cfg.Valid(), IsNil, Commentf("%s is invalid coin", item))
		cfg1 := NewAdminConfig(GetAdminConfigKey(item), "1233", addr)
		c.Check(cfg1.Valid(), NotNil, Commentf("%s is not valid coin", item))
	}
	adminCfg := NewAdminConfig(GSLKey, "100", addr)
	c.Check(adminCfg.Empty(), Equals, false)
	c.Check(adminCfg.DbKey(), Equals, "GSL_"+addr.String())
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
			key:     MinValidatorBondKey,
			value:   "blab",
		},
		{
			address: addr,
			key:     PoolRefundGasKey,
			value:   "whatever",
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
		GSLKey:               "0.3",
		StakerAmtIntervalKey: "100",
		MinValidatorBondKey:  sdk.NewUint(common.One * 10).String(),
		WhiteListGasAssetKey: "1000bep",
		PoolRefundGasKey:     strconv.Itoa(common.One / 10),
		DefaultPoolStatus:    "Enabled",
	}
	for k, v := range input {
		if k.Default() != v {
			c.Errorf("expected: %s , however we got: %s", v, k.Default())
		}
	}
}
