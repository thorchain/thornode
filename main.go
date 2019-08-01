package main

import "gitlab.com/thorchain/bepswap/observe/x/silverback"

func main() {
	binance := silverback.NewBinance()
	
	silverback.SyncBal(*binance)
	silverback.NewServer(*binance).Start()
	silverback.NewClient(*binance).Start()
}
