package main

import (
	"flag"
	"fmt"
	"log"

	sdk "github.com/binance-chain/go-sdk/client"
	"github.com/binance-chain/go-sdk/keys"

	"gitlab.com/thorchain/bepswap/thor-node/test/smoke"
)

// main : Generate our pool address.
func main() {
	apiAddr := flag.String("a", "testnet-dex.binance.org", "Binance API Address.")
	network := flag.Int("n", 0, "The network to use.")
	addrType := flag.String("t", "MASTER", "The type [POOL|MASTER].")
	flag.Parse()

	n := smoke.NewNetwork(*network)
	keyManager, _ := keys.NewKeyManager()
	if _, err := sdk.NewDexClient(*apiAddr, n.Type, keyManager); nil != err {
		log.Fatalf("%v", err)
	}

	fmt.Printf("export %v=%v\n", *addrType, keyManager.GetAddr())
	privKey, _ := keyManager.ExportAsPrivateKey()
	fmt.Printf("export %v_KEY=%v\n", *addrType, privKey)
}
