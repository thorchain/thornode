package config

import (
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gitlab.com/thorchain/bepswap/common"
)

// Configuration values
type Configuration struct {
	PoolAddress           common.BnbAddress         `json:"pool_address" mapstructure:"pool_address"`
	DEXHost               string                    `json:"dex_host" mapstructure:"dex_host"`
	MessageProcessor      int                       `json:"message_processor" mapstructure:"message_processor"`
	ObserverDbPath        string                    `json:"observer_db_path" mapstructure:"observer_db_path"`
	BlockScanner          BlockScannerConfiguration `json:"block_scanner" mapstructure:"block_scanner"`
	StateChain            StateChainConfiguration   `json:"state_chain" mapstructure:"state_chain"`
	ObserverRetryInterval time.Duration             `json:"observer_retry_interval" mapstructure:"observer_retry_interval"`
	Metric                MetricConfiguration       `json:"metric" mapstructure:"metric"`
}

// BlockScannerConfiguration settings for BlockScanner
type BlockScannerConfiguration struct {
	RPCHost                    string        `json:"rpc_host" mapstructure:"rpc_host"`
	Scheme                     string        `json:"scheme"  mapstructure:"scheme"`
	StartBlockHeight           int64         `json:"start_block_height" mapstructure:"start_block_height"`
	BlockScanProcessors        int           `json:"block_scan_processors" mapstructure:"block_scan_processors"`
	HttpRequestTimeout         time.Duration `json:"http_request_timeout" mapstructure:"http_request_timeout"`
	HttpRequestReadTimeout     time.Duration `json:"http_request_read_timeout" mapstructure:"http_request_read_timeout"`
	HttpRequestWriteTimeout    time.Duration `json:"http_request_write_timeout" mapstructure:"http_request_write_timeout"`
	MaxHttpRequestRetry        int           `json:"max_http_request_retry" mapstructure:"max_http_request_retry"`
	BlockHeightDiscoverBackoff time.Duration `json:"block_height_discover_back_off" mapstructure:"block_height_discover_back_off"`
	BlockRetryInterval         time.Duration `json:"block_retry_interval" mapstructure:"block_retry_interval"`
}

// StateChainConfiguration
type StateChainConfiguration struct {
	ChainID         string `json:"chain_id" mapstructure:"chain_id" `
	ChainHost       string `json:"chain_host" mapstructure:"chain_host"`
	ChainHomeFolder string `json:"chain_home_folder" mapstructure:"chain_home_folder"`
	SignerName      string `json:"signer_name" mapstructure:"signer_name"`
	SignerPasswd    string `json:"signer_passwd" mapstructure:"signer_passwd"`
}
type MetricConfiguration struct {
	Enabled      bool          `json:"enabled" mapstructure:"enabled"`
	ListenPort   int           `json:"listen_port" mapstructure:"listen_port"`
	ReadTimeout  time.Duration `json:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" mapstructure:"write_timeout"`
}

func applyDefaultObserverConfig() {
	viper.SetDefault("dexhost", "testnet-dex.binance.org")
	viper.SetDefault("message_processor", "10")
	viper.SetDefault("observer_db_path", "observer_data")
	viper.SetDefault("observer_retry_interval", "2s")
	applyBlockScannerDefault()
	viper.SetDefault("state_chain.chain_id", "statechain")
	viper.SetDefault("state_chain.chain_host", "localhost:1317")
	viper.SetDefault("metric.listen_port", "9000")
	viper.SetDefault("metric.read_timeout", "30s")
	viper.SetDefault("metric.write_timeout", "30s")
}

// LoadObserveConfig
func LoadObserverConfig(file string) (*Configuration, error) {
	applyDefaultObserverConfig()
	var cfg Configuration
	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Dir(file))
	viper.SetConfigName(strings.TrimRight(path.Base(file), ".json"))
	if err := viper.ReadInConfig(); nil != err {
		return nil, errors.Wrap(err, "fail to read from config file")
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.Unmarshal(&cfg); nil != err {
		return nil, errors.Wrap(err, "fail to unmarshal")
	}
	return &cfg, nil
}

// SignerConfiguration all the configures need by signer
type SignerConfiguration struct {
	SignerDbPath     string                    `json:"signer_db_path" mapstructure:"signer_db_path"`
	MessageProcessor int                       `json:"message_processor" mapstructure:"message_processor"`
	BlockScanner     BlockScannerConfiguration `json:"block_scanner" mapstructure:"block_scanner"`
	Binance          BinanceConfiguration      `json:"binance" mapstructure:"binance"`
	StateChain       StateChainConfiguration   `json:"state_chain" mapstructure:"state_chain"`
	RetryInterval    time.Duration             `json:"retry_interval" mapstructure:"retry_interval"`
	Metric           MetricConfiguration       `json:"metric" mapstructure:"metric"`
}

// BinanceConfiguration all the configurations for binance client
type BinanceConfiguration struct {
	DEXHost    string `json:"dex_host" mapstructure:"dex_host"`
	PrivateKey string `json:"private_key" mapstructure:"private_key"`
}

func applyBlockScannerDefault() {
	viper.SetDefault("block_scanner.start_block_height", "0")
	viper.SetDefault("block_scanner.scheme", "https")
	viper.SetDefault("block_scanner.block_scan_processors", "2")
	viper.SetDefault("block_scanner.http_request_timeout", "30s")
	viper.SetDefault("block_scanner.http_request_read_timeout", "30s")
	viper.SetDefault("block_scanner.http_request_write_timeout", "30s")
	viper.SetDefault("block_scanner.max_http_request_retry", "10")
	viper.SetDefault("block_scanner.block_height_discover_back_off", "1s")
	viper.SetDefault("block_scanner.block_retry_interval", "1s")
}
func applyDefaultSignerConfig() {
	viper.SetDefault("signer_db_path", "signer_db")
	viper.SetDefault("message_processor", "10")
	applyBlockScannerDefault()
	viper.SetDefault("state_chain.chain_host", "localhost:1317")
	viper.SetDefault("retry_interval", "2s")
	viper.SetDefault("metric.listen_port", "9000")
	viper.SetDefault("metric.read_timeout", "30s")
	viper.SetDefault("metric.write_timeout", "30s")
}

// LoadObserveConfig
func LoadSignerConfig(file string) (*SignerConfiguration, error) {
	applyDefaultSignerConfig()
	var cfg SignerConfiguration
	viper.AddConfigPath(".")
	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Dir(file))
	viper.SetConfigName(strings.TrimRight(path.Base(file), ".json"))

	if err := viper.ReadInConfig(); nil != err {
		return nil, errors.Wrap(err, "fail to read from config file")
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.Unmarshal(&cfg); nil != err {
		return nil, errors.Wrap(err, "fail to unmarshal")
	}
	return &cfg, nil
}
