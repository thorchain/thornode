package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
)

// Settings for the service
type Settings struct {
	Pools      []string `json:"pools"`
	Port       int      `json:"port"`
	DexBaseUrl string   `json:"dex_base_url"`
	IsTestNet  bool     `json:"is_test_net"`
}

// LoadFromFile load the config from file
func LoadFromFile(filePathName string) (*Settings, error) {
	body, err := ioutil.ReadFile(filePathName)
	if nil != err {
		return nil, errors.Wrapf(err, "fail to read %s", filePathName)
	}
	var cfg Settings
	if err := json.Unmarshal(body, &cfg); nil != err {
		return nil, errors.Wrapf(err, "fail to deserialize settings,file path name : %s", filePathName)
	}
	return &cfg, nil
}
