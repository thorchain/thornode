package common

import (
	"os"
	"strings"
)

// ChainNetwork is to indicate which chain environment THORNode are working with
type ChainNetwork uint8

const (
	// TestNetwork for test
	TestNet ChainNetwork = iota
	// ProdNetwork for main net
	MainNet
)

// GetCurrentChainNetwork determinate what kind of network currently it is working with
func GetCurrentChainNetwork() ChainNetwork {
	if strings.EqualFold(os.Getenv("NET"), "testnet") || strings.EqualFold(os.Getenv("NET"), "testnet") {
		return TestNet
	}
	return MainNet
}
