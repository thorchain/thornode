package main

import (
	c "gitlab.com/thorchain/bepswap/observe/common"
	"gitlab.com/thorchain/bepswap/observe/x/observer"
	"gitlab.com/thorchain/bepswap/observe/x/signer"
)

func main() {
	txChan := make(chan []byte)

	observer.NewObserver(txChan).Start()
	signer.NewSigner(txChan).Start()

	c.StartWebServer()
}
