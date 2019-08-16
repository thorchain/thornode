package main

import (
	"gitlab.com/thorchain/bepswap/observe/x/observer"
)

func main() {
	observer.NewObserver().Start()
}
