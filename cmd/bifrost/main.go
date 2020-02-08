package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"

	"gitlab.com/thorchain/thornode/bifrost/pkg/chainclients"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/observer"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	"gitlab.com/thorchain/thornode/bifrost/signer"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/cmd"
	"gitlab.com/thorchain/thornode/common"
)

// THORNode define version / revision here , so THORNode could inject the version from CI pipeline if THORNode want to
var (
	version  string
	revision string
)

const (
	serverIdentity      = "bifrost"
)

func printVersion() {
	fmt.Printf("%s v%s, rev %s\n", serverIdentity, version, revision)
}

func main() {
	showVersion := flag.Bool("version", false, "Shows version")
	logLevel := flag.StringP("log-level", "l", "info", "Log Level")
	pretty := flag.BoolP("pretty-log", "p", false, "Enables unstructured prettified logging. This is useful for local debugging")
	cfgFile := flag.StringP("cfg", "c", "config", "configuration file with extension")
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

	if len(cfg.Chains) == 0 {
		log.Fatal().Err(err).Msg("missing chains")
		return
	}

	chains := chainclients.LoadChains(thorKeys, cfg.Chains, cfg.TSS, thorchainBridge)

	// start observer
	obs, err := observer.NewObserver(cfg.Observer, thorchainBridge, pubkeyMgr, chains[common.BNBChain], m)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to create observer")
	}
	obs.Start();

	// start signer
	sign, err := signer.NewSigner(cfg.Signer, thorchainBridge, thorKeys, pubkeyMgr, cfg.TSS, chains, m)
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
}
