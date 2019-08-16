package main

import (
	"gitlab.com/thorchain/bepswap/observe/x/signer"
)

func main() {
	signer.NewSigner().Start()
}
