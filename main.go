package main

import (
	"os"
	"time"
	silverback "gitlab.com/thorchain/bepswap/observe/x/silverback"
)

func main() {
	port := os.Getenv("PORT")
	server := silverback.NewServer(port)
	server.Start()

	poolAddress := os.Getenv("POOL_ADDRESS")
	dexHost := os.Getenv("DEX_HOST")
	client := silverback.NewClient(30 * time.Second, poolAddress, dexHost)
	client.Start()
}
