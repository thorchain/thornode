package main

import (
	"flag"
	"strings"

	"gitlab.com/thorchain/bepswap/thor-node/test/smoke"
)

func main() {
	apiAddr := flag.String("a", "testnet-dex.binance.org", "Binance API Address.")
	masterKey := flag.String("m", "", "The master key of the wallet to transfer assets to.")
	keyList := flag.String("k", "", "A comma separated list of keys to hoover assets from.")
	network := flag.Int("n", 0, "The network to use.")
	debug := flag.Bool("d", false, "Enable debugging of the Binance transactions.")
	flag.Parse()

	keys := strings.Split(*keyList, ",")

	h := smoke.NewSweep(*apiAddr, *masterKey, keys, *network, *debug)
	h.EmptyWallets()
}
