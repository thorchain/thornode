package main

import (
	"flag"
	"strings"

	btypes "github.com/binance-chain/go-sdk/common/types"
	"gitlab.com/thorchain/bepswap/thornode/test/smoke"
)

func main() {
	apiAddr := flag.String("a", "testnet-dex.binance.org", "Binance API Address.")
	masterKey := flag.String("m", "", "The master key of the wallet to transfer assets to.")
	keyList := flag.String("k", "", "A comma separated list of keys to hoover assets from.")
	network := flag.Int("n", 0, "The network to use.")
	debug := flag.Bool("d", false, "Enable debugging of the Binance transactions.")
	flag.Parse()

	keys := strings.Split(*keyList, ",")

	net := btypes.TestNetwork
	if *network > 0 {
		net = btypes.ProdNetwork
	}

	h := smoke.NewSweep(*apiAddr, *masterKey, keys, net, *debug)
	h.EmptyWallets()
}
