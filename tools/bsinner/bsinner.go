package main

import (
	"flag"
	"log"
	"os"

	"gitlab.com/thorchain/thornode/test/smoke"
)

func main() {
	apiAddr := flag.String("a", "https://data-seed-pre-0-s3.binance.org/", "Binance RPC address.")
	faucetKey := flag.String("f", "", "The faucet private key.")
	poolKey := flag.String("k", "", "The pool key.")
	environment := flag.String("e", "local", "The environment to use [local|staging|develop|production]. Defaults to local")
	bal := flag.String("b", "", "Balances json file")
	txns := flag.String("t", "", "Transactions json file")
	fastFail := flag.Bool("x", false, "Enable fast fail")
	debug := flag.Bool("d", false, "Enable debugging of the Binance transactions.")
	flag.Parse()

	if *txns == "" {
		log.Fatal("No transactions json file")
	}

	if *bal == "" {
		log.Fatal("No balances json file")
	}

	s := smoke.NewSmoke(*apiAddr, *faucetKey, *poolKey, *environment, *bal, *txns, *fastFail, *debug)
	successful := s.Run()
	if successful {
		os.Exit(0)
	}
	os.Exit(1)
}
