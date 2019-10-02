package main

import (
	"flag"

	"gitlab.com/thorchain/bepswap/statechain/x/smoke"
	"gitlab.com/thorchain/bepswap/statechain/x/smoke/types"
)

func main() {
	masterKey := flag.String("m", "", "The master private key.")
	poolKey := flag.String("p", "", "The pool private key.")
	config := flag.String("c", types.DefaultConfig, "Path to the config file.")
	flag.Parse()

	s := smoke.NewSmoke(*masterKey, *poolKey, *config)
	s.Run()
}
