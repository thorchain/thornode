package types

import "fmt"

type AdminConfig struct {
	Key   string `json:"k"`
	Value string `json:"v"`
}

func NewAdminConfig(k, v string) AdminConfig {
	return AdminConfig{k, v}
}

func (c AdminConfig) Empty() bool {
	return c.Key == ""
}

func (c AdminConfig) String() string {
	return fmt.Sprintf("Config: %s --> %s", c.Key, c.Value)
}
