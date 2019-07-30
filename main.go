package main

import (
	"os"
	"time"
	silverback "gitlab.com/thorchain/bepswap/observe/x/silverback"
	rest "gitlab.com/thorchain/bepswap/observe/x/rest"
)

func main() {
	port := os.Getenv("PORT")
	rest.Start(port)
	
	poolAddress := os.Getenv("POOL_ADDRESS")
	dexHost := os.Getenv("DEX_HOST")
	silverback.Start(30 * time.Second, poolAddress, dexHost)
}
