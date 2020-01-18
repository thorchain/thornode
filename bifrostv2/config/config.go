package config

import (
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Configuration values
type Configuration struct {
	Thorchain ClientConfiguration     `json:"thorchain" mapstructure:"thorchain"`
	Metric    MetricConfiguration     `json:"metric" mapstructure:"metric"`
	Chains    []ChainConfigurations   `json:"chains" mapstructure:"chains"`
	TxScanner TxScannerConfigurations `json:"tx_scanner" mapstructure:"tx_scanner"`
	TxSigner  TxSignerConfigurations  `json:"tx_signer" mapstructure:"tx_signer"`
	BackOff   BackOff                 `json:"back_off" mapstructure:"back_off"`
}

type TxScannerConfigurations struct {
	BlockChains []ChainConfigurations
}

type TxSignerConfigurations struct {
	BlockChains  []ChainConfigurations
	BlockScanner BlockScannerConfiguration `json:"block_scanner" mapstructure:"block_scanner"`
}

type BackOff struct {
	InitialInterval     time.Duration `json:"initial_interval" mapstructure:"initial_interval"`
	RandomizationFactor float64       `json:"randomization_factor" mapstructure:"randomization_factor"`
	Multiplier          float64       `json:"multiplier" mapstructure:"multiplier"`
	MaxInterval         time.Duration `json:"max_interval" mapstructure:"max_interval"`
	MaxElapsedTime      time.Duration `json:"max_elapsed_time" mapstructure:"max_elapsed_time"`
}

type ChainConfigurations struct {
	Name         string `json:"name" mapstructure:"name"`
	Enabled      bool   `json:"enabled" mapstructure:"enabled"`
	ChainHost    string `json:"chain_host" mapstructure:"chain_host"`
	ChainNetwork string `json:"chain_network" mapstructure:"chain_network"`
	UserName     string `json:"username" mapstructure:"username"`
	Password     string `json:"password" mapstructure:"password"`
	HTTPostMode  bool   `json:"http_post_mode" mapstructure:"http_post_mode"` // Bitcoin core only supports HTTP POST mode
	DisableTLS   bool   `json:"disable_tls" mapstructure:"disable_tls"`       // Bitcoin core does not provide TLS by default
	BackOff      BackOff
}

// ClientConfiguration for http thorchain client
type ClientConfiguration struct {
	ChainID         string `json:"chain_id" mapstructure:"chain_id" `
	ChainHost       string `json:"chain_host" mapstructure:"chain_host"`
	ChainHomeFolder string `json:"chain_home_folder" mapstructure:"chain_home_folder"`
	SignerName      string `json:"signer_name" mapstructure:"signer_name"`
	SignerPasswd    string `json:"signer_passwd" mapstructure:"signer_passwd"`
	BackOff         BackOff
}

// BlockScannerConfiguration settings for BlockScanner
type BlockScannerConfiguration struct {
	RPCHost                    string        `json:"rpc_host" mapstructure:"rpc_host"`
	StartBlockHeight           int64         `json:"-"`
	BlockScanProcessors        int           `json:"block_scan_processors" mapstructure:"block_scan_processors"`
	HttpRequestTimeout         time.Duration `json:"http_request_timeout" mapstructure:"http_request_timeout"`
	HttpRequestReadTimeout     time.Duration `json:"http_request_read_timeout" mapstructure:"http_request_read_timeout"`
	HttpRequestWriteTimeout    time.Duration `json:"http_request_write_timeout" mapstructure:"http_request_write_timeout"`
	MaxHttpRequestRetry        int           `json:"max_http_request_retry" mapstructure:"max_http_request_retry"`
	BlockHeightDiscoverBackoff time.Duration `json:"block_height_discover_back_off" mapstructure:"block_height_discover_back_off"`
	BlockRetryInterval         time.Duration `json:"block_retry_interval" mapstructure:"block_retry_interval"`
	EnforceBlockHeight         bool          `json:"enforce_block_height" mapstructure:"enforce_block_height"`
}

// MetricConfiguration hold config for prometheus metrics
type MetricConfiguration struct {
	Enabled      bool          `json:"enabled" mapstructure:"enabled"`
	ListenPort   int           `json:"listen_port" mapstructure:"listen_port"`
	ReadTimeout  time.Duration `json:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" mapstructure:"write_timeout"`
}

// TSSConfiguration hold TSS config
type TSSConfiguration struct {
	Scheme string `json:"scheme" mapstructure:"scheme"`
	Host   string `json:"host" mapstructure:"host"`
	Port   int    `json:"port" mapstructure:"port"`
	NodeId string `json:"node_id" mapstructure:"node_id"`
}

func applyDefaultConfig() {
	viper.SetDefault("metric.listen_port", "9000")
	viper.SetDefault("metric.read_timeout", "30s")
	viper.SetDefault("metric.write_timeout", "30s")
	viper.SetDefault("back_off.initial_interval", 500*time.Millisecond)
	viper.SetDefault("back_off.randomization_factor", 0.5)
	viper.SetDefault("back_off.multiplier", 1.5)
	viper.SetDefault("back_off.max_interval", 3*time.Minute)
	viper.SetDefault("back_off.max_elapsed_time", 168*time.Hour) // 7 days. Due to node sync time's being so random
	viper.SetDefault("tx_signer.block_scanner.start_block_height", "0")
	viper.SetDefault("tx_signer.block_scanner.block_scan_processors", "2")
	viper.SetDefault("tx_signer.block_scanner.http_request_timeout", "30s")
	viper.SetDefault("tx_signer.block_scanner.http_request_read_timeout", "30s")
	viper.SetDefault("tx_signer.block_scanner.http_request_write_timeout", "30s")
	viper.SetDefault("tx_signer.block_scanner.max_http_request_retry", "10")
	viper.SetDefault("tx_signer.block_scanner.block_height_discover_back_off", "1s")
	viper.SetDefault("tx_signer.block_scanner.block_retry_interval", "1s")
}

func LoadBiFrostConfig(file string) (*Configuration, error) {
	applyDefaultConfig()
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

	// set global backoff settings to all chains config.
	for _, chain := range cfg.Chains {
		chain.BackOff = cfg.BackOff
		cfg.TxScanner.BlockChains = append(cfg.TxScanner.BlockChains, chain)
		cfg.TxSigner.BlockChains = append(cfg.TxSigner.BlockChains, chain)
	}

	return &cfg, nil
}
