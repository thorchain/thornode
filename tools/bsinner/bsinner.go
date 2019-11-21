package main

import (
	"flag"
	"log"

	btypes "github.com/binance-chain/go-sdk/common/types"
	"gitlab.com/thorchain/bepswap/thornode/test/smoke"
)

// smoke test run a json config file that is a series of transaction and expected results.
func main() {
	apiAddr := flag.String("a", "testnet-dex.binance.org", "Binance API Address.")
	faucetKey := flag.String("f", "", "The faucet private key.")
	poolAddr := flag.String("p", "", "The pool address.")
	poolKey := flag.String("k", "", "The pool key.")
	environment := flag.String("e", "stage", "The environment to use [local|staging|develop|production].")
	config := flag.String("c", "", "Path to the config file.")
	network := flag.Int("n", 0, "The network to use.")
	sweep := flag.Bool("s", false, "Sweep funds back on exit [Default: false]")
	logFile := flag.String("l", "/tmp/smoke.json", "The path to the log file [/tmp/smoke.json].")
	debug := flag.Bool("d", false, "Enable debugging of the Binance transactions.")
	flag.Parse()

	if *faucetKey == "" {
		log.Fatal("No faucet key set!")
	}

	if *poolAddr == "" && *poolKey == "" {
		log.Fatal("No pool address or pool key set!")
	}

	if *config == "" {
		log.Fatal("No config file provided!")
	}

	net := btypes.TestNetwork
	if *network > 0 {
		net = btypes.ProdNetwork
	}

	s := smoke.NewSmoke(*apiAddr, *faucetKey, *poolAddr, *poolKey, *environment, *config, net, *logFile, *sweep, *debug)
	s.Run()
}
