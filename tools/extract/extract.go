package main

import (
	"flag"
	"fmt"
	"log"

	sdk "github.com/binance-chain/go-sdk/client"
	"github.com/binance-chain/go-sdk/keys"

	ctypes "github.com/binance-chain/go-sdk/common/types"
)

// main : Extract information from a Binance keystore file.
func main() {
	apiAddr := flag.String("a", "testnet-dex.binance.org", "Binance API Address.")
	network := flag.Int("n", 0, "The network to use.")
	addrType := flag.String("t", "MASTER", "The type [POOL|MASTER].")
	file := flag.String("f", "", "Path to the keystore file.")
	password := flag.String("p", "", "Password for the keystore file.")
	flag.Parse()

	n := ctypes.TestNetwork
	if *network > 0 {
		n = ctypes.ProdNetwork
	}

	keyManager, err := keys.NewKeyStoreKeyManager(*file, *password)
	if err != nil {
		log.Panic(err)
	}

	if _, err := sdk.NewDexClient(*apiAddr, n, keyManager); nil != err {
		log.Panic(err)
	}

	fmt.Printf("export %v_ADDR=%v\n", *addrType, keyManager.GetAddr())
	privKey, _ := keyManager.ExportAsPrivateKey()
	fmt.Printf("export %v_KEY=%v\n", *addrType, privKey)
}
