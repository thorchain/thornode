package types

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
)

type AdminConfigKey string

const (
	UnknownKey                AdminConfigKey = "Unknown"
	GSLKey                    AdminConfigKey = "GSL"
	StakerAmtIntervalKey      AdminConfigKey = "StakerAmtInterval"
	MinValidatorBondKey       AdminConfigKey = "MinValidatorBond"
	WhiteListGasTokenKey      AdminConfigKey = "WhiteListGasToken"      // How much gas token we mint and send it to the newly whitelisted bep address
	DesireValidatorSetKey     AdminConfigKey = "DesireValidatorSet"     // how much validators we would like to have
	RotatePerBlockHeightKey   AdminConfigKey = "RotatePerBlockHeight"   // how many blocks we try to rotate validators
	ValidatorsChangeWindowKey AdminConfigKey = "ValidatorsChangeWindow" // when should we open the rotate window, nominate validators, and identify who should be out
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
	case string(GSLKey):
		return GSLKey
	case string(StakerAmtIntervalKey):
		return StakerAmtIntervalKey
	case string(MinValidatorBondKey):
		return MinValidatorBondKey
	case string(WhiteListGasTokenKey):
		return WhiteListGasTokenKey
	case string(DesireValidatorSetKey):
		return DesireValidatorSetKey
	case string(RotatePerBlockHeightKey):
		return RotatePerBlockHeightKey
	case string(ValidatorsChangeWindowKey):
		return ValidatorsChangeWindowKey
	default:
		return UnknownKey
	}
}

func (k AdminConfigKey) Default() string {
	switch k {
	case GSLKey:
		return "0.3"
	case StakerAmtIntervalKey:
		return "100"
	case MinValidatorBondKey:
		return sdk.NewUint(common.One * 10).String()
	case WhiteListGasTokenKey:
		return "1000bep"
	case DesireValidatorSetKey:
		return "4"
	case RotatePerBlockHeightKey:
		return "28800" // a day
	case ValidatorsChangeWindowKey:
		return "1200" // one hour
	default:
		return ""
	}
}

// Ensure the value for a given key is a valid
func (k AdminConfigKey) ValidValue(value string) error {
	var err error
	switch k {
	case GSLKey, StakerAmtIntervalKey:
		_, err = common.NewAmount(value)
	case MinValidatorBondKey:
		_, err = sdk.ParseUint(value)
	case WhiteListGasTokenKey:
		_, err = sdk.ParseCoins(value)
	case DesireValidatorSetKey, RotatePerBlockHeightKey, ValidatorsChangeWindowKey: // int64
		_, err = strconv.ParseInt(value, 10, 64)
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
