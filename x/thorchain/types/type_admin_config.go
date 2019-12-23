package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AdminConfigKey string

const (
	UnknownKey        AdminConfigKey = "Unknown"
	DefaultPoolStatus AdminConfigKey = "DefaultPoolStatus" // When a pool get created automatically , what status do THORNode set it in
)

func (k AdminConfigKey) String() string {
	return string(k)
}

func (k AdminConfigKey) IsValidKey() bool {
	key := GetAdminConfigKey(k.String())
	return key != UnknownKey
}
func GetAdminConfigKey(key string) AdminConfigKey {
	switch key {
	case string(DefaultPoolStatus):
		return DefaultPoolStatus
	default:
		return UnknownKey
	}
}

func (k AdminConfigKey) Default() string {
	switch k {
	case DefaultPoolStatus:
		return "Enabled"
	default:
		return ""
	}
}

// Ensure the value for a given key is a valid
func (k AdminConfigKey) ValidValue(value string) error {
	var err error
	switch k {
	case DefaultPoolStatus:
		if GetPoolStatus(value) == Suspended {
			return errors.New("invalid pool status")
		}
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

	if !c.Key.IsValidKey() {
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
