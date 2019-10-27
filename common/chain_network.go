package common

import (
	"os"
	"strings"
)

type ChainNetwork uint8

const (
	TestNetwork ChainNetwork = iota
	ProdNetwork
)

// GetCurrentChainNetwork determinate what kind of network currently it is working with
func GetCurrentChainNetwork() ChainNetwork {
	if strings.EqualFold(os.Getenv("NET"), "testnet") {
		return TestNetwork
	}
	return ProdNetwork
}
