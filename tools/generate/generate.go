package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"

	bech32 "github.com/btcsuite/btcutil/bech32"
)

// main : Generate our pool address.
func main() {
	network := flag.Int("n", 0, "The network to use.")
	addrType := flag.String("t", "MASTER", "The type [POOL|MASTER].")
	flag.Parse()

	types.Network = types.TestNetwork
	if *network > 0 {
		types.Network = types.ProdNetwork
	}
	keyManager, err := keys.NewKeyManager()
	if err != nil {
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

	mnem, err := keyManager.ExportAsMnemonic()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	fmt.Printf("export %v_MNEMONIC=%v\n", *addrType, mnem)
}
