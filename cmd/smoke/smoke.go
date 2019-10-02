package main

import (
	"log"
	"flag"

	"gitlab.com/thorchain/bepswap/statechain/x/smoke"
	"gitlab.com/thorchain/bepswap/statechain/x/smoke/types"
)

func main() {
	apiAddr := flag.String("a", "testnet-dex.binance.org", "Binance API Address.")
	masterKey := flag.String("m", "", "The master private key.")
	poolKey := flag.String("p", "", "The pool private key.")
	config := flag.String("c", types.DefaultConfig, "Path to the config file.")
	network := flag.Int("n", 0, "The network to use.")
	flag.Parse()

	if *masterKey == "" {
		log.Fatal("No Master Key set!")
	}

	if *poolKey == "" {
		log.Fatal("No Pool Key set!")
	}

	s := smoke.NewSmoke(*apiAddr, *masterKey, *poolKey, *config, *network)
	s.Run()
}
