package config

import (
	"time"

	"gitlab.com/thorchain/bepswap/common"
)

// Configuration values
type Configuration struct {
	PoolAddress               common.BnbAddress         `json:"pool_address" env:"POOL_ADDRESS" required:"true"`
	DEXHost                   string                    `json:"dex_host" env:"DEX_HOST"`
	MessageProcessor          int                       `json:"message_processor" default:"10"`
	ObserverDbPath            string                    `json:"observer_db_path" env:"LEVEL_DB_OBSERVER_PATH"`
	BlockScannerConfiguration BlockScannerConfiguration `json:"block_scanner_configuration"`
	StateChainConfiguration   StateChainConfiguration   `json:"state_chain_configuration"`
	ObserverRetryInterval     time.Duration             `json:"observer_retry_interval" default:"120s" env:"OBSERVER_RETRY_INTERVAL"`
}

// BlockScannerConfiguration settings for BlockScanner
type BlockScannerConfiguration struct {
	RPCHost                    string        `json:"rpc_host" env:"RPC_HOST"`
	Scheme                     string        `json:"scheme" default:"https" env:"RPC_SCHEME"`
	StartBlockHeight           int64         `json:"start_block_height" env:"START_BLOCK_HEIGHT"`
	BlockScanProcessors        int           `json:"block_scan_processors" env:"BLOCK_SCAN_PROCESSORS"`
	HttpRequestTimeout         time.Duration `json:"http_request_timeout" default:"10s"`
	HttpRequestReadTimeout     time.Duration `json:"http_request_read_timeout" default:"30s"`
	HttpRequestWriteTimeout    time.Duration `json:"http_request_write_timeout" default:"30s"`
	MaxHttpRequestRetry        int           `json:"max_http_request_retry" env:"BLOCK_SCAN_MAX_HTTP_RETRY" default:"10"`
	BlockHeightDiscoverBackoff time.Duration `json:"block_height_discover_back_off" default:"1s"`
	BlockRetryInterval         time.Duration `json:"block_retry_interval" default:"5m"`
}

// StateChainConfiguration
type StateChainConfiguration struct {
	ChainID         string `json:"chain_id" env:"CHAIN_ID"`
	ChainHost       string `json:"chain_host" env:"CHAIN_HOST"`
	ChainHomeFolder string `json:"chain_home_folder" env:"CHAIN_HOME_FOLDER"`
	SignerName      string `json:"signer_name" env:"SIGNER_NAME"`
	SignerPasswd    string `json:"signer_passwd" env:"SIGNER_PASSWD"`
}

// SignerConfiguration all the configures need by signer
type SignerConfiguration struct {
	SignerDbPath              string                    `json:"signer_db_path" env:"SIGNER_DB_PATH" default:"signer_db" required:"true"`
	MessageProcessor          int                       `json:"message_processor" default:"10" env:"SIGNER_MESSAGE_PROCESSORS"`
	BlockScannerConfiguration BlockScannerConfiguration `json:"block_scanner_configuration"`
	Binance                   BinanceConfiguration      `json:"binance"`
	StateChainConfiguration   StateChainConfiguration   `json:"state_chain_configuration"`
	RetryInterval             time.Duration             `json:"retry_interval" default:"2s" env:"SIGNER_RETRY_INTERVAL"`
}

// BinanceConfiguration all the configurations for binance client
type BinanceConfiguration struct {
	DEXHost    string `json:"dex_host" required:"true" env:"DEX_HOST"`
	PrivateKey string `json:"private_key" required:"true" env:"BINANCE_PRIVATE_KEY"`
}
