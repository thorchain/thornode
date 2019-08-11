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

func (c AdminConfig) Valid() error {
	if c.Key == "" {
		return fmt.Errorf("Key cannot be empty")
	}
	if c.Value == "" {
		return fmt.Errorf("Value cannot be empty")
	}
	return nil
}

func (c AdminConfig) String() string {
	return fmt.Sprintf("Config: %s --> %s", c.Key, c.Value)
}
