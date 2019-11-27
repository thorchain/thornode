package main

import (
	"flag"
	"strings"

	"gitlab.com/thorchain/thornode/test/smoke"
)

func main() {
	apiAddr := flag.String("a", "data-seed-pre-0-s3.binance.org", "Binance RPC Address.")
	masterKey := flag.String("m", "", "The master key of the wallet to transfer assets to.")
	keyList := flag.String("k", "", "A comma separated list of keys to hoover assets from.")
	debug := flag.Bool("d", false, "Enable debugging of the Binance transactions.")
	flag.Parse()

	keys := strings.Split(*keyList, ",")

	h := smoke.NewSweep(*apiAddr, *masterKey, keys, *debug)
	h.EmptyWallets()
}
