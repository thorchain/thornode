package types

import (
	"fmt"

	common "gitlab.com/thorchain/bepswap/common"
)

type AdminConfigKey string

const (
	UnknownKey           AdminConfigKey = "Unknown"
	GSLKey               AdminConfigKey = "GSL"
	TSLKey               AdminConfigKey = "TSL"
	StakerAmtIntervalKey AdminConfigKey = "StakerAmtInterval"
	PoolAddressKey       AdminConfigKey = "PoolAddress"
	MRRAKey              AdminConfigKey = `MRRA` // MRRA means MinimumRefundRuneAmount, if the tx send to pool has less then this amount of RUNE , we are not going to refund it
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
	default:
		return UnknownKey
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
	}
	return err
}

type AdminConfig struct {
	Key   AdminConfigKey `json:"key"`
	Value string         `json:"value"`
}

func NewAdminConfig(key AdminConfigKey, value string) AdminConfig {
	return AdminConfig{
		Key:   key,
		Value: value,
	}
}

func (c AdminConfig) Empty() bool {
	return c.Key == ""
}

func (c AdminConfig) Valid() error {
	if c.Key == "" {
		return fmt.Errorf("Key cannot be empty")
	}
	if c.Key == UnknownKey {
		return fmt.Errorf("Key not recognized")
	}
	if c.Value == "" {
		return fmt.Errorf("Value cannot be empty")
	}
	if err := c.Key.ValidValue(c.Value); err != nil {
		return err
	}
	return nil
}

func (c AdminConfig) String() string {
	return fmt.Sprintf("Config: %s --> %s", c.Key, c.Value)
}
