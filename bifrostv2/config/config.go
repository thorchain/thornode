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
	ThorChain ThorChainConfiguration  `json:"thorchain" mapstructure:"thorchain"`
	Metric    MetricConfiguration     `json:"metric" mapstructure:"metric"`
	Chains    ChainsConfigurations    `json:"chains" mapstructure:"chains"`
	TxScanner TxScannerConfigurations `json:"tx_scanner" mapstructure:"tx_scanner"`
	TxSigner  TxSignerConfigurations  `json:"tx_signer" mapstructure:"tx_signer"`
}

type TxScannerConfigurations struct {
	BlockChains *ChainsConfigurations
}

type TxSignerConfigurations struct {
	BlockChains *ChainsConfigurations
}

type ChainsConfigurations struct {
	BTC BTCConfiguration `json:"btc" mapstructure:"btc"`
	ETH ETHConfiguration `json:"eth" mapstructure:"eth"`
	BNB BNBConfiguration `json:"bnb" mapstructure:"bnb"`
	XMR XMRConfiguration `json:"xmr" mapstructure:"xmr"`
}

type CommonBlockChainConfigurations struct {
	Enabled      bool   `json:"enabled" mapstructure:"enabled"`
	ChainHost    string `json:"chain_host" mapstructure:"chain_host"`
	ChainNetwork string `json:"chain_network" mapstructure:"chain_network"`
	UserName     string `json:"username" mapstructure:"username"`
	Password     string `json:"password" mapstructure:"password"`
}

type BTCConfiguration struct {
	CommonBlockChainConfigurations `mapstructure:",squash"`
	HTTPostMode                    bool `json:"http_post_mode" mapstructure:"http_post_mode"` // Bitcoin core only supports HTTP POST mode
	DisableTLS                     bool `json:"disable_tls" mapstructure:"disable_tls"`       // Bitcoin core does not provide TLS by default
}

type ETHConfiguration struct {
	CommonBlockChainConfigurations `mapstructure:",squash"`
}

// BNBConfiguration all the configurations for binance client
type BNBConfiguration struct {
	CommonBlockChainConfigurations `mapstructure:",squash"`
}

type XMRConfiguration struct {
	CommonBlockChainConfigurations `mapstructure:",squash"`
}

// ThorChainConfiguration
type ThorChainConfiguration struct {
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

func applyDefaultConfig() {
	viper.SetDefault("metric.listen_port", "9000")
	viper.SetDefault("metric.read_timeout", "30s")
	viper.SetDefault("metric.write_timeout", "30s")
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

	// Set TxScanner and Signer to use the global blockchain settings so no dups is needed.
	cfg.TxScanner.BlockChains = &cfg.Chains
	cfg.TxSigner.BlockChains = &cfg.Chains
	return &cfg, nil
}

// SignerConfiguration all the configures need by signer
type SignerConfiguration struct {
	SignerDbPath     string `json:"signer_db_path" mapstructure:"signer_db_path"`
	MessageProcessor int    `json:"message_processor" mapstructure:"message_processor"`
	// BlockScanner     BlockScannerConfiguration `json:"block_scanner" mapstructure:"block_scanner"`
	Binance       BNBConfiguration       `json:"binance" mapstructure:"binance"`
	StateChain    ThorChainConfiguration `json:"state_chain" mapstructure:"state_chain"`
	RetryInterval time.Duration          `json:"retry_interval" mapstructure:"retry_interval"`
	Metric        MetricConfiguration    `json:"metric" mapstructure:"metric"`
	KeySign       TSSConfiguration       `json:"key_sign" mapstructure:"key_sign"`
	UseTSS        bool                   `json:"use_tss" mapstructure:"use_tss"`
	KeyGen        TSSConfiguration       `json:"key_gen" mapstructure:"key_gen"`
}

// TSSConfiguration
type TSSConfiguration struct {
	Scheme string `json:"scheme" mapstructure:"scheme"`
	Host   string `json:"host" mapstructure:"host"`
	Port   int    `json:"port" mapstructure:"port"`
	NodeId string `json:"node_id" mapstructure:"node_id"`
}
