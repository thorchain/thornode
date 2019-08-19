package config

import (
	"os"
	"time"

	"gitlab.com/thorchain/bepswap/common"
)

// Configuration values
type Configuration struct {
	PoolAddress               common.BnbAddress         `json:"pool_address" env:"POOL_ADDRESS" required:"true"`
	RuneAddress               string                    `json:"rune_address" env:"RUNE_ADDRESS"`
	DEXHost                   string                    `json:"dex_host" env:"DEX_HOST"`
	RPCHost                   string                    `json:"rpc_host" env:"RPC_HOST"`
	PrivateKey                string                    `json:"private_key" env:"PRIVATE_KEY"`
	ChainHost                 string                    `json:"chain_host" env:"CHAIN_HOST"`
	SignerName                string                    `json:"signer_name" env:"SIGNER_NAME"`
	SignerPasswd              string                    `json:"signer_passwd" env:"SIGNER_PASSWD"`
	SignerDbPath              string                    `json:"signer_db_path" env:"LEVEL_DB_SIGNER_PATH"`
	SocketPoing               time.Duration             `json:"socket_poing" default:"30s"`
	MessageProcessor          int                       `json:"message_processor" default:"10"`
	BlockScannerConfiguration BlockScannerConfiguration `json:"block_scanner_configuration"`
}

// BlockScannerConfiguration settings for BlockScanner
type BlockScannerConfiguration struct {
	RPCHost                    string        `json:"rpc_host" env:"RPC_HOST"`
	ObserverDbPath             string        `json:"observer_db_path" env:"LEVEL_DB_OBSERVER_PATH"`
	StartBlockHeight           int64         `json:"start_block_height" env:"START_BLOCK_HEIGHT"`
	BlockScanProcessors        int           `json:"block_scan_processors" env:"BLOCK_SCAN_PROCESSORS"`
	HttpRequestTimeout         time.Duration `json:"http_request_timeout" default:"10s"`
	HttpRequestReadTimeout     time.Duration `json:"http_request_read_timeout" default:"30s"`
	HttpRequestWriteTimeout    time.Duration `json:"http_request_write_timeout" default:"30s"`
	MaxHttpRequestRetry        int           `json:"max_http_request_retry" env:"BLOCK_SCAN_MAX_HTTP_RETRY" default:"10"`
	BlockHeightDiscoverBackoff time.Duration `json:"block_height_discover_back_off"`
	BlockRetryInterval         time.Duration `json:"block_retry_interval" default:"5m"`
}


// TODO to be removed later
var (
	PoolAddress = os.Getenv("POOL_ADDRESS")
	//RuneAddress    = os.Getenv("RUNE_ADDRESS")
	DEXHost   = os.Getenv("DEX_HOST")
	RPCHost   = os.Getenv("RPC_HOST")
	PrivKey   = os.Getenv("PRIVATE_KEY")
	ChainHost = os.Getenv("CHAIN_HOST")
	//SignerPasswd   = os.Getenv("SIGNER_PASSWD")
	//ObserverDbPath = os.Getenv("LEVEL_DB_OBSERVER_PATH")
	SignerDbPath = os.Getenv("LEVEL_DB_SIGNER_PATH")
)
