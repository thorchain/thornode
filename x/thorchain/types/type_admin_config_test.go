package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type AdminConfigSuite struct{}

var _ = Suite(&AdminConfigSuite{})

func (s AdminConfigSuite) TestGetKey(c *C) {
	keys := []string{
		"Unknown",
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
	uintAmnt := []string{"PoolRefundGas"}
addr := GetRandomBech32Addr()
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
		WhiteListGasAssetKey: "1000bep",
		PoolRefundGasKey:     strconv.Itoa(common.One / 10),
		DefaultPoolStatus:    "Enabled",
	}
	for k, v := range input {
		if k.Default() != v {
			c.Errorf("expected: %s , however THORNode got: %s", v, k.Default())
		}
	}
}
