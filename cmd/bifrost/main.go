package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	btsskeygen "github.com/binance-chain/tss-lib/ecdsa/keygen"
	sdk "github.com/cosmos/cosmos-sdk/types"
	golog "github.com/ipfs/go-log"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
	"gitlab.com/thorchain/tss/go-tss/common"
	"gitlab.com/thorchain/tss/go-tss/tss"

	app "gitlab.com/thorchain/thornode"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/observer"
	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients"
	"gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	"gitlab.com/thorchain/thornode/bifrost/signer"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/cmd"
)

// THORNode define version / revision here , so THORNode could inject the version from CI pipeline if THORNode want to
var (
	version  string
	revision string
)

const (
	serverIdentity = "bifrost"
)

func printVersion() {
	fmt.Printf("%s v%s, rev %s\n", serverIdentity, version, revision)
}

func main() {
	showVersion := flag.Bool("version", false, "Shows version")
	logLevel := flag.StringP("log-level", "l", "info", "Log Level")
	pretty := flag.BoolP("pretty-log", "p", false, "Enables unstructured prettified logging. This is useful for local debugging")
	cfgFile := flag.StringP("cfg", "c", "config", "configuration file with extension")
	tssPreParam := flag.StringP("preparm", "t", "", "pre-generated PreParam file used for tss")
	flag.Parse()

	if *showVersion {
		printVersion()
		return
	}

	initPrefix()
	initLog(*logLevel, *pretty)

	// load configuration file
	cfg, err := config.LoadBiFrostConfig(*cfgFile)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to load config ")
	}
	cfg.Thorchain.SignerPasswd = os.Getenv("SIGNER_PASSWD")

	// metrics
	m, err := metrics.NewMetrics(cfg.Metrics)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to create metric instance")
	}
	if err := m.Start(); err != nil {
		log.Fatal().Err(err).Msg("fail to start metric collector")
	}

	// thorchain bridge
	thorchainBridge, err := thorclient.NewThorchainBridge(cfg.Thorchain, m)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to create new thorchain bridge")
	}
	if err := thorchainBridge.EnsureNodeWhitelistedWithTimeout(); err != nil {
		log.Fatal().Err(err).Msg("node account is not whitelisted, can't start")
	}

	// PubKey Manager
	pubkeyMgr, err := pubkeymanager.NewPubKeyManager(cfg.Thorchain.ChainHost, m)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to create pubkey manager")
	}
	if err := pubkeyMgr.Start(); err != nil {
		log.Fatal().Err(err).Msg("fail to start pubkey manager")
	}

	// get thorchain key manager
	thorKeys, err := thorclient.NewKeys(cfg.Thorchain.ChainHomeFolder, cfg.Thorchain.SignerName, cfg.Thorchain.SignerPasswd)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to load keys")
	}

	// setup TSS signing
	priKey, err := thorKeys.GetPrivateKey()
	if err != nil {
		log.Fatal().Err(err).Msg("fail to get private key")
	}
	bootstrapPeers, err := cfg.TSS.GetBootstrapPeers()
	if err != nil {
		log.Fatal().Err(err).Msg("fail to get bootstrap peers")
	}
	tssIns, err := tss.NewTss(bootstrapPeers,
		cfg.TSS.P2PPort,
		priKey,
		cfg.TSS.Rendezvous,
		app.DefaultCLIHome,
		common.TssConfig{
			KeyGenTimeout:   30 * time.Second,
			KeySignTimeout:  10 * time.Second,
			PreParamTimeout: 5 * time.Minute,
		}, getLocalPreParam(*tssPreParam))
	if err != nil {
		log.Fatal().Err(err).Msg("fail to create tss instance")
	}

	if err := tssIns.Start(); err != nil {
		log.Err(err).Msg("fail to start tss instance")
	}

	healthServer := NewHealthServer(cfg.TSS.InfoAddress, tssIns)
	go func() {
		defer log.Info().Msg("health server exit")
		if err := healthServer.Start(); err != nil {
			log.Error().Err(err).Msg("fail to start health server")
		}
	}()
	if len(cfg.Chains) == 0 {
		log.Fatal().Err(err).Msg("missing chains")
		return
	}

	// ensure we have a protocol for chain RPC Hosts
	for _, chainCfg := range cfg.Chains {
		if len(chainCfg.RPCHost) == 0 {
			log.Fatal().Err(err).Msg("missing chain RPC host")
			return
		}
		if !strings.HasPrefix(chainCfg.RPCHost, "http") {
			chainCfg.RPCHost = fmt.Sprintf("http://%s", chainCfg.RPCHost)
		}

		if len(chainCfg.BlockScanner.RPCHost) == 0 {
			log.Fatal().Err(err).Msg("missing chain RPC host")
			return
		}
		if !strings.HasPrefix(chainCfg.BlockScanner.RPCHost, "http") {
			chainCfg.BlockScanner.RPCHost = fmt.Sprintf("http://%s", chainCfg.BlockScanner.RPCHost)
		}
	}

	chains := chainclients.LoadChains(thorKeys, cfg.Chains, tssIns, thorchainBridge, m)

	// start observer
	obs, err := observer.NewObserver(pubkeyMgr, chains, thorchainBridge, m)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to create observer")
	}
	if err = obs.Start(); err != nil {
		log.Fatal().Err(err).Msg("fail to start observer")
	}

	// start signer
	sign, err := signer.NewSigner(cfg.Signer, thorchainBridge, thorKeys, pubkeyMgr, tssIns, cfg.TSS, chains, m)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to create instance of signer")
	}
	if err := sign.Start(); err != nil {
		log.Fatal().Err(err).Msg("fail to start signer")
	}

	// wait....
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Info().Msg("stop signal received")

	// stop observer
	if err := obs.Stop(); err != nil {
		log.Fatal().Err(err).Msg("fail to stop observer")
	}
	// stop signer
	if err := sign.Stop(); err != nil {
		log.Fatal().Err(err).Msg("fail to stop signer")
	}
	// stop go tss
	tssIns.Stop()
	if err := healthServer.Stop(); err != nil {
		log.Fatal().Err(err).Msg("fail to stop health server")
	}
}

func initPrefix() {
	cosmosSDKConfg := sdk.GetConfig()
	cosmosSDKConfg.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	cosmosSDKConfg.Seal()
}

func initLog(level string, pretty bool) {
	l, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Warn().Msgf("%s is not a valid log-level, falling back to 'info'", level)
	}
	var out io.Writer = os.Stdout
	if pretty {
		out = zerolog.ConsoleWriter{Out: os.Stdout}
	}
	zerolog.SetGlobalLevel(l)
	log.Logger = log.Output(out).With().Str("service", serverIdentity).Logger()

	logLevel := golog.LevelInfo
	switch l {
	case zerolog.DebugLevel:
		logLevel = golog.LevelDebug
	case zerolog.InfoLevel:
		logLevel = golog.LevelInfo
	case zerolog.ErrorLevel:
		logLevel = golog.LevelError
	case zerolog.FatalLevel:
		logLevel = golog.LevelFatal
	case zerolog.PanicLevel:
		logLevel = golog.LevelPanic
	}
	golog.SetAllLoggers(logLevel)
	if err := golog.SetLogLevel("tss-lib", level); err != nil {
		log.Fatal().Err(err).Msg("fail to set tss-lib loglevel")
	}
}

func getLocalPreParam(file string) *btsskeygen.LocalPreParams {
	if len(file) == 0 {
		return nil
	}
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal().Msgf("fail to read file:%s", file)
		return nil
	}
	buf = bytes.Trim(buf, "\n")
	log.Info().Msg(string(buf))
	result, err := hex.DecodeString(string(buf))
	if err != nil {
		log.Fatal().Msg("fail to hex decode the file content")
		return nil
	}
	var preParam btsskeygen.LocalPreParams
	if err := json.Unmarshal(result, &preParam); err != nil {
		log.Fatal().Msg("fail to unmarshal file content to LocalPreParams")
		return nil
	}
	return &preParam
}
