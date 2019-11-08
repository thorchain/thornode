package main

import (
	"flag"
	"fmt"
	"log"

	sdk "github.com/binance-chain/go-sdk/client"
	"github.com/binance-chain/go-sdk/keys"

	bech32 "github.com/btcsuite/btcutil/bech32"

	"gitlab.com/thorchain/bepswap/thornode/test/smoke"
)

// main : Generate our pool address.
func main() {
	apiAddr := flag.String("a", "testnet-dex.binance.org", "Binance API Address.")
	network := flag.Int("n", 0, "The network to use.")
	addrType := flag.String("t", "MASTER", "The type [POOL|MASTER].")
	flag.Parse()

	n := smoke.NewNetwork(*network)
	keyManager, err := keys.NewKeyManager()
	if err != nil {
		log.Fatalf("%v", err)
	}

	if _, err := sdk.NewDexClient(*apiAddr, n.Type, keyManager); nil != err {
		log.Fatalf("%v", err)
	}

	fmt.Printf("export %v=%v\n", *addrType, keyManager.GetAddr())
	privKey, err := keyManager.ExportAsPrivateKey()
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Printf("export %v_KEY=%v\n", *addrType, privKey)

	keyBytes := keyManager.GetPrivKey().PubKey().Bytes()
	conv, _ := bech32.ConvertBits(keyBytes, 8, 5, true)
	pubKey, _ := bech32.Encode("bnbp", conv)
	fmt.Printf("export %v_PUBKEY=%v\n", *addrType, pubKey)
}
