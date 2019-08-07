package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	configName = `statechain`
)

var (
	globalSettings = DefaultSettings()
)

// Settings for the service
type Settings struct {
	// TODO to be removed
	Pools      []string `json:"pools"`
	Port       int      `json:"port"`
	DexBaseUrl string   `json:"dex_base_url"`
	IsTestNet  bool     `json:"is_test_net"`

	GlobalPoolSlip       float64 `json:"global_pool_slip"`
	GlobalTradeSlipLimit float64 `json:"global_trade_slip_limit"`
	// TODO add stake limit here
}

// DefaultSettings
func DefaultSettings() *Settings {
	return &Settings{
		GlobalPoolSlip:       0.2,
		GlobalTradeSlipLimit: 0.1,
	}
}

func GetGlobalSettings() *Settings {
	return globalSettings
}

// LoadFromFile load the config from file
func LoadFromFile(homeFolder string) (*Settings, error) {
	viper.SetConfigName(configName)
	viper.AddConfigPath(homeFolder)
	if err := viper.ReadInConfig(); nil != err {
		// write the default to file
		s := DefaultSettings()
		buf, err := json.Marshal(s)
		if nil != err {
			return nil, errors.Wrap(err, "fail to marshal the default setting")
		}
		if err := ioutil.WriteFile(filepath.Join(homeFolder, "config/statechain.json"), buf, 0400); nil != err {
			return nil, errors.Wrap(err, "fail to write to file")
		}
		return s, nil
	}
	if err := viper.Unmarshal(globalSettings); nil != err {
		return nil, errors.Wrap(err, "fail to unmarshal config")
	}
	return globalSettings, nil
}
