package main

import "gitlab.com/thorchain/bepswap/observe/x/silverback"

func main() {
	binance := silverback.NewBinance()
	pool := silverback.NewPool(binance.PoolAddress)
	
	silverback.SyncBal(*binance)
	silverback.NewServer(*binance, *pool).Start()
	silverback.NewClient(*binance, *pool).Start()
}
