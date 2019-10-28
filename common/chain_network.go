package common

import (
	"os"
	"strings"
)

// ChainNetwork is to indicate which chain environment we are working with
type ChainNetwork uint8

const (
	// TestNetwork for test
	TestNetwork ChainNetwork = iota
	// ProdNetwork for main net
	ProdNetwork
)

// GetCurrentChainNetwork determinate what kind of network currently it is working with
func GetCurrentChainNetwork() ChainNetwork {
	if strings.EqualFold(os.Getenv("NET"), "testnet") {
		return TestNetwork
	}
	return ProdNetwork
}
