package types

import "fmt"

type AdminConfig struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewAdminConfig(key, value string) AdminConfig {
	return AdminConfig{
		Key:   key,
		Value: value,
	}
}

func (c AdminConfig) Empty() bool {
	return c.Key == ""
}

func (c AdminConfig) String() string {
	return fmt.Sprintf("Config: %s --> %s", c.Key, c.Value)
}
