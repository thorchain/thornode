package types

import "os"

var (
	PoolAddress		= os.Getenv("POOL_ADDRESS")
	RuneAddress		= os.Getenv("RUNE_ADDRESS")
	DEXHost				= os.Getenv("DEX_HOST")
	RPCHost				= os.Getenv("RPC_HOST")
	PrivKey				= os.Getenv("PRIVATE_KEY")
	ChainHost			= os.Getenv("CHAIN_HOST")
	SignerPasswd	= os.Getenv("SIGNER_PASSWD")
	RedisUrl 			= os.Getenv("REDIS_URL")
	RedisPasswd		= os.Getenv("REDIS_PASSWORD")
	StatusPort		= os.Getenv("PORT")
)

