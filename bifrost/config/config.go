package config

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Configuration struct {
	Observer  ObserverConfiguration `json:"observer" mapstructure:"observer"`
	Signer    SignerConfiguration   `json:"signer" mapstructure:"signer"`
	Thorchain ClientConfiguration   `json:"thorchain" mapstructure:"thorchain"`
	Metrics   MetricsConfiguration  `json:"metrics" mapstructure:"metrics"`
	Chains    []ChainConfiguration  `json:"chains" mapstructure:"chains"`
	TSS       TSSConfiguration      `json:"tss" mapstructure:"tss"`
	BackOff   BackOff               `json:"back_off" mapstructure:"back_off"`
}

// ObserverConfiguration values
type ObserverConfiguration struct {
	ObserverDbPath string                    `json:"observer_db_path" mapstructure:"observer_db_path"`
	BlockScanner   BlockScannerConfiguration `json:"block_scanner" mapstructure:"block_scanner"`
	RetryInterval  time.Duration             `json:"retry_interval" mapstructure:"retry_interval"`
}

// SignerConfiguration all the configures need by signer
type SignerConfiguration struct {
	SignerDbPath  string                    `json:"signer_db_path" mapstructure:"signer_db_path"`
	BlockScanner  BlockScannerConfiguration `json:"block_scanner" mapstructure:"block_scanner"`
	RetryInterval time.Duration             `json:"retry_interval" mapstructure:"retry_interval"`
}

// BackOff configuration
type BackOff struct {
	InitialInterval     time.Duration `json:"initial_interval" mapstructure:"initial_interval"`
	RandomizationFactor float64       `json:"randomization_factor" mapstructure:"randomization_factor"`
	Multiplier          float64       `json:"multiplier" mapstructure:"multiplier"`
	MaxInterval         time.Duration `json:"max_interval" mapstructure:"max_interval"`
	MaxElapsedTime      time.Duration `json:"max_elapsed_time" mapstructure:"max_elapsed_time"`
}

// ChainConfiguration configuration
type ChainConfiguration struct {
	Name         string `json:"name" mapstructure:"name"`
	ChainHost    string `json:"chain_host" mapstructure:"chain_host"`
	ChainNetwork string `json:"chain_network" mapstructure:"chain_network"`
	UserName     string `json:"username" mapstructure:"username"`
	Password     string `json:"password" mapstructure:"password"`
	RPCHost      string `jsonn:"rpc_host" mapstructure:"rpc_host"`
	HTTPostMode  bool   `json:"http_post_mode" mapstructure:"http_post_mode"` // Bitcoin core only supports HTTP POST mode
	DisableTLS   bool   `json:"disable_tls" mapstructure:"disable_tls"`       // Bitcoin core does not provide TLS by default
	BackOff      BackOff
}

// TSSConfiguration
type TSSConfiguration struct {
	Scheme string `json:"scheme" mapstructure:"scheme"`
	Host   string `json:"host" mapstructure:"host"`
	Port   int    `json:"port" mapstructure:"port"`
	NodeId string `json:"node_id" mapstructure:"node_id"`
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
	ChainID                    string        `json:"chain_id" mapstructure:"chain_id"`
}

// ClientConfiguration
type ClientConfiguration struct {
	ChainID         string `json:"chain_id" mapstructure:"chain_id" `
	ChainHost       string `json:"chain_host" mapstructure:"chain_host"`
	ChainHomeFolder string `json:"chain_home_folder" mapstructure:"chain_home_folder"`
	SignerName      string `json:"signer_name" mapstructure:"signer_name"`
	SignerPasswd    string `json:"signer_passwd" mapstructure:"signer_passwd"`
	BackOff         BackOff
}

type MetricsConfiguration struct {
	Enabled      bool          `json:"enabled" mapstructure:"enabled"`
	ListenPort   int           `json:"listen_port" mapstructure:"listen_port"`
	ReadTimeout  time.Duration `json:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" mapstructure:"write_timeout"`
	Chains       []string      `json:"chains" mapstructure:"chains"`
}

func LoadBiFrostConfig(file string) (*Configuration, error) {
	applyDefaultConfig()
	var cfg Configuration
	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Dir(file))
	viper.SetConfigName(strings.TrimRight(path.Base(file), ".json"))
	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "fail to read from config file")
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "fail to unmarshal")
	}

	// Set global backoff settings to all chains config. Maybe is pointless and inefficient as we have it in global config
	for i, chain := range cfg.Chains {
		cfg.Chains[i].BackOff = cfg.BackOff
		cfg.Metrics.Chains = append(cfg.Metrics.Chains, chain.Name)
	}

	return &cfg, nil
}

func applyDefaultConfig() {
	viper.SetDefault("metrics.listen_port", "9000")
	viper.SetDefault("metrics.read_timeout", "30s")
	viper.SetDefault("metrics.write_timeout", "30s")
	viper.SetDefault("metrics.chains", []string{"BNB"})
	viper.SetDefault("thorchain.chain_id", "thorchain")
	viper.SetDefault("thorchain.chain_host", "localhost:1317")
	viper.SetDefault("back_off.initial_interval", 500*time.Millisecond)
	viper.SetDefault("back_off.randomization_factor", 0.5)
	viper.SetDefault("back_off.multiplier", 1.5)
	viper.SetDefault("back_off.max_interval", 3*time.Minute)
	viper.SetDefault("back_off.max_elapsed_time", 168*time.Hour) // 7 days. Due to node sync time's being so random
	applyDefaultObserverConfig()
	applyDefaultSignerConfig()
}

func applyBlockScannerDefault(path string) {
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.start_block_height", path), "0")
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.block_scan_processors", path), "2")
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.http_request_timeout", path), "30s")
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.http_request_read_timeout", path), "30s")
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.http_request_write_timeout", path), "30s")
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.max_http_request_retry", path), "10")
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.block_height_discover_back_off", path), "1s")
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.block_retry_interval", path), "1s")
}

func applyDefaultObserverConfig() {
	viper.SetDefault("observer.observer_db_path", "observer_data")
	viper.SetDefault("observer.retry_interval", "2s")
	applyBlockScannerDefault("observer")
	viper.SetDefault("observer.block_scanner.chain_id", "BNB")
}

func applyDefaultSignerConfig() {
	viper.SetDefault("signer.signer_db_path", "signer_db")
	applyBlockScannerDefault("signer")
	viper.SetDefault("signer.retry_interval", "2s")
	viper.SetDefault("signer.block_scanner.chain_id", "ThorChain")
}
