package main

import (
	"gitlab.com/thorchain/bepswap/observe/x/observer"
	"gitlab.com/thorchain/bepswap/observe/x/signer"
	common "gitlab.com/thorchain/bepswap/observe/common"	
)

func main() {
	observer.NewObserver().Start()
	signer.NewSigner().Start()

	common.StartWebServer()
}
