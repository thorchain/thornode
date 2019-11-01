package smoke

import (
	ctypes "github.com/cbarraford/go-sdk/common/types"
	btypes "github.com/cbarraford/go-sdk/types"
)

// Network is the different between testnet and mainNet
type Network struct {
	Type    ctypes.ChainNetwork
	ChainID string
}

func NewNetwork(network int) Network {
	return Network{
		Type:    networkType(network),
		ChainID: chainID(network),
	}
}

func networkType(network int) ctypes.ChainNetwork {
	if network == 0 {
		return ctypes.TestNetwork
	} else {
		return ctypes.ProdNetwork
	}
}

func chainID(network int) string {
	if network == 0 {
		return btypes.TestnetChainID
	} else {
		return btypes.ProdChainID
	}
}
