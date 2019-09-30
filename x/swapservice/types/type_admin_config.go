package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
)

type AdminConfigKey string

const (
	UnknownKey           AdminConfigKey = "Unknown"
	GSLKey               AdminConfigKey = "GSL"
	TSLKey               AdminConfigKey = "TSL"
	StakerAmtIntervalKey AdminConfigKey = "StakerAmtInterval"
	PoolAddressKey       AdminConfigKey = "PoolAddress"
	PoolExpiryKey        AdminConfigKey = "PoolExpiry"
	MinStakerCoinsKey    AdminConfigKey = "MinStakerCoins"
	MRRAKey              AdminConfigKey = "MRRA" // MRRA means MinimumRefundRuneAmount, if the tx send to pool has less then this amount of RUNE , we are not going to refund it
	MinValidatorBondKey  AdminConfigKey = "MinValidatorBond"
	WhiteListGasTokenKey AdminConfigKey = "WhiteListGasToken" // How much gas token we mint and send it to the newly whitelisted bep address
)

func (k AdminConfigKey) String() string {
	return string(k)
}

func GetAdminConfigKey(key string) AdminConfigKey {
	switch key {
	case string(GSLKey):
		return GSLKey
	case string(TSLKey):
		return TSLKey
	case string(StakerAmtIntervalKey):
		return StakerAmtIntervalKey
	case string(PoolAddressKey):
		return PoolAddressKey
	case string(PoolExpiryKey):
		return PoolExpiryKey
	case string(MinStakerCoinsKey):
		return MinStakerCoinsKey
	case string(MRRAKey):
		return MRRAKey
	case string(MinValidatorBondKey):
		return MinValidatorBondKey
	case string(WhiteListGasTokenKey):
		return WhiteListGasTokenKey
	default:
		return UnknownKey
	}
}

func (k AdminConfigKey) Default() string {
	switch k {
	case GSLKey:
		return "0.3"
	case TSLKey:
		return "0.1"
	case StakerAmtIntervalKey:
		return "100"
	case MinStakerCoinsKey:
		return "1bep"
	case MRRAKey:
		return sdk.NewUint(common.One).String()
	case MinValidatorBondKey:
		return sdk.NewUint(common.One * 10).String()
	case WhiteListGasTokenKey:
		return "1000bep"
	default:
		return ""
	}
}

// Ensure the value for a given key is a valid
func (k AdminConfigKey) ValidValue(value string) error {
	var err error
	switch k {
	case GSLKey, TSLKey, StakerAmtIntervalKey:
		_, err = common.NewAmount(value)
	case PoolAddressKey:
		_, err = common.NewBnbAddress(value)
	case MRRAKey, MinValidatorBondKey:
		_, err = sdk.ParseUint(value)
	case MinStakerCoinsKey, WhiteListGasTokenKey:
		_, err = sdk.ParseCoins(value)
	}
	return err
}

type AdminConfig struct {
	Key     AdminConfigKey `json:"key"`
	Value   string         `json:"value"`
	Address sdk.AccAddress `json:"address"`
}

func NewAdminConfig(key AdminConfigKey, value string, address sdk.AccAddress) AdminConfig {
	return AdminConfig{
		Key:     key,
		Value:   value,
		Address: address,
	}
}

func (c AdminConfig) Empty() bool {
	return c.Key == ""
}

func (c AdminConfig) Valid() error {
	if c.Address.Empty() {
		return fmt.Errorf("address cannot be empty")
	}
	if c.Key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	if c.Key == UnknownKey {
		return fmt.Errorf("key not recognized")
	}
	if c.Value == "" {
		return fmt.Errorf("value cannot be empty")
	}
	if err := c.Key.ValidValue(c.Value); err != nil {
		return err
	}
	return nil
}

func (c AdminConfig) DbKey() string {
	return fmt.Sprintf("%s_%s", c.Key.String(), c.Address.String())
}

func (c AdminConfig) String() string {
	return fmt.Sprintf("Config: %s --> %s", c.Key, c.Value)
}
