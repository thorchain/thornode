package main

import (
	"flag"
	"fmt"

	sdk "github.com/binance-chain/go-sdk/client"
	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
)

// main : Generate our pool address.
func main() {
	apiAddr := flag.String("a", "testnet-dex.binance.org", "Binance API Address.")
	network := flag.Uint("n", 0, "The network to use.")
	flag.Parse()

	sdk.NewDexClient(*apiAddr, selectedNet(network), keyManager)
	privKey, _ := keyManager.ExportAsPrivateKey()
	fmt.Printf("%v\n", privKey)
}

// selectedNet : Get the Binance network type
func selectedNet(network int) types.ChainNetwork {
	if network == 0 {
		return types.TestNetwork
	} else {
		return types.ProdNetwork
	}
}
