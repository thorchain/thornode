package main

import (
	"os"

	"gitlab.com/thorchain/bepswap/observe/x/observer"
	"gitlab.com/thorchain/bepswap/observe/x/signer"
)

func main() {
	chainHost := os.Getenv("CHAIN_HOST")
	poolAddress := os.Getenv("POOL_ADDRESS")
	dexHost := os.Getenv("DEX_HOST")
	rpcHost := os.Getenv("RPC_HOST")
	runeAddress := os.Getenv("RUNE_ADDRESS")

	txChan := make(chan []byte)

	observer.NewObserver(poolAddress, dexHost, rpcHost, chainHost, runeAddress, txChan).Start()
	signer.NewSigner(poolAddress, dexHost, chainHost, txChan).Start()

	observer.StartWebServer()
}
