package main

import (
	"flag"
	"log"

	"gitlab.com/thorchain/bepswap/thor-node/test/smoke"
)

// smoke test run a json config file that is a series of transaction and expected results.
func main() {
	apiAddr := flag.String("a", "testnet-dex.binance.org", "Binance API Address.")
	bankKey := flag.String("b", "", "The bank private key.")
	poolKey := flag.String("p", "", "The pool private key.")
	environment := flag.String("e", "stage", "The environment to use [local|stage|dev|prod].")
	config := flag.String("c", "", "Path to the config file.")
	network := flag.Int("n", 0, "The network to use.")
	debug := flag.Bool("d", false, "Enable debugging of the Binance transactions.")
	flag.Parse()

	if *bankKey == "" {
		log.Fatal("No bank key set!")
	}

	if *poolKey == "" {
		log.Fatal("No pool key set!")
	}

	if *config == "" {
		log.Fatal("No config file provided!")
	}

	s := smoke.NewSmoke(*apiAddr, *bankKey, *poolKey, *environment, *config, *network, *debug)
	s.Run()
}
