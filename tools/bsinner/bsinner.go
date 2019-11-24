package main

import (
	"flag"
	"log"
	"os"

	"gitlab.com/thorchain/bepswap/thornode/test/smoke"
)

// smoke test run a json config file that is a series of transaction and expected results.
func main() {
	apiAddr := flag.String("a", "http://data-seed-pre-0-s3.binance.org/", "Binance RPC Address.")
	faucetKey := flag.String("f", "", "The faucet private key.")
	poolKey := flag.String("k", "", "The pool key.")
	environment := flag.String("e", "stage", "The environment to use [local|staging|develop|production].")
	bal := flag.String("b", "", "Balances json file")
	txns := flag.String("t", "", "Transactions json file")
	sweep := flag.Bool("s", false, "Sweep funds back on exit [Default: false]")
	fastFail := flag.Bool("x", false, "Enable fast fail")
	debug := flag.Bool("d", false, "Enable debugging of the Binance transactions.")
	flag.Parse()

	if *txns == "" {
		log.Fatal("No transactions json file")
	}

	if *bal == "" {
		log.Fatal("No balances json file")
	}

	if *faucetKey == "" {
		log.Fatal("No faucet key set!")
	}

	s := smoke.NewSmoke(*apiAddr, *faucetKey, *poolKey, *environment, *bal, *txns, *sweep, *fastFail, *debug)
	successful := s.Run()
	if successful {
		os.Exit(0)
	}
	os.Exit(1)
}
