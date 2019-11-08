package main

import (
	"flag"
	"log"

	"gitlab.com/thorchain/bepswap/thornode/test/smoke"
)

// smoke test run a json config file that is a series of transaction and expected results.
func main() {
	apiAddr := flag.String("a", "testnet-dex.binance.org", "Binance API Address.")
	faucetKey := flag.String("f", "", "The faucet private key.")
	poolKey := flag.String("p", "", "The pool private key.")
	environment := flag.String("e", "stage", "The environment to use [local|staging|develop|production].")
	config := flag.String("c", "", "Path to the config file.")
	network := flag.Int("n", 0, "The network to use.")
	resultsFile := flag.String("l", "/tmp/smoke.json", "Where test results will be saved [/tmp/smoke.json].")
	thorchainFile := flag.String("t", "/tmp/thorchain.json", "Where Thorchain state results will be saved [/tmp/thorchain.json].")
	debug := flag.Bool("d", false, "Enable debugging of the Binance transactions.")
	flag.Parse()

	if *faucetKey == "" {
		log.Fatal("No faucet key set!")
	}

	if *poolKey == "" {
		log.Fatal("No pool key set!")
	}

	if *config == "" {
		log.Fatal("No config file provided!")
	}

	s := smoke.NewSmoke(*apiAddr, *faucetKey, *poolKey, *environment, *config, *network, *resultsFile, *thorchainFile, *debug)
	s.Run()
}
