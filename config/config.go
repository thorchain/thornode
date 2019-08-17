package types

import (
	"os"
	"time"
)

// TODO: Create config struct that reads from config file
// TODO: validate configuration on startup, panic if invalid

var (
	PoolAddress    = os.Getenv("POOL_ADDRESS")
	RuneAddress    = os.Getenv("RUNE_ADDRESS")
	DEXHost        = os.Getenv("DEX_HOST")
	RPCHost        = os.Getenv("RPC_HOST")
	PrivKey        = os.Getenv("PRIVATE_KEY")
	ChainHost      = os.Getenv("CHAIN_HOST")
	SignerPasswd   = os.Getenv("SIGNER_PASSWD")
	ObserverDbPath = os.Getenv("LEVEL_DB_OBSERVER_PATH")
	SignerDbPath   = os.Getenv("LEVEL_DB_SIGNER_PATH")
)

var (
	SocketPong = 30 * time.Second
)
