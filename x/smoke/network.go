package smoke

import (
	btypes "github.com/binance-chain/go-sdk/types"
	ctypes "github.com/binance-chain/go-sdk/common/types"
)

// selectedNet : Get the Binance network type
func selectedNet(network int) (ctypes.ChainNetwork, string) {
	if network == 0 {
		return ctypes.TestNetwork, btypes.TestnetChainID
	} else {
		return ctypes.ProdNetwork, btypes.ProdChainID
	}
}
