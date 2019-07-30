package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	app "github.com/jpthor/cosmos-swap"
	"github.com/jpthor/cosmos-swap/config"
	"github.com/jpthor/cosmos-swap/exchange"
	"github.com/jpthor/cosmos-swap/storage"
)

var (
	version  string
	revision string
)

const (
	serverIdentity = "service"
)

func exchangeCmd() *cobra.Command {
	exCmd := &cobra.Command{
		Use:   "exchange",
		Short: "Integration with Binance exchange",
		Long:  "exchange",
		RunE:  startExchange,
	}
	exCmd.Flags().BoolP("version", "v", false, "show current version")
	exCmd.Flags().StringP("log-level", "l", "info", "Log Level")
	exCmd.Flags().BoolP("pretty-log", "p", false, "Enables unstructured prettified logging. This is useful for local debugging")
	exCmd.Flags().StringP("dir", "d", "data", "data folder where we put level db")
	exCmd.Flags().StringP("cfg", "c", "config.json", "configuration file path name")
	return flags.PostCommands(
		exCmd,
	)[0]
}

func printVersion() {
	fmt.Printf("%s v%s, rev %s\n", "service", version, revision)
}

func startExchange(cmd *cobra.Command, args []string) error {
	showVersion, err := cmd.Flags().GetBool("version")
	if nil != err {
		return errors.Wrap(err, "fail to get version flag")
	}
	if showVersion {
		printVersion()
		return nil
	}
	logLevel, err := cmd.Flags().GetString("log-level")
	if nil != err {
		return errors.Wrap(err, "fail to get log-level from flag")
	}
	pretty, err := cmd.Flags().GetBool("pretty-log")
	if nil != err {
		return errors.Wrap(err, "fail to get pretty-log value from flag")
	}
	initLog(logLevel, pretty)
	dir, err := cmd.Flags().GetString("dir")
	if nil != err {
		return errors.Wrap(err, "fail to get data folder from flag")
	}
	cfg, err := cmd.Flags().GetString("cfg")
	if nil != err {
		return errors.Wrap(err, "fail to get config file path name from flag")
	}
	s, err := config.LoadFromFile(cfg)
	if nil != err {
		return errors.Wrapf(err, "fail to load config from %s", cfg)
	}
	ds, err := storage.NewDataStore(dir, log.Logger)
	if nil != err {
		return errors.Wrapf(err, "fail to create data storage")
	}
	ws, err := exchange.NewWallets(ds, log.Logger)
	if nil != err {
		return errors.Wrapf(err, "fail to create wallets")
	}

	clictx := context.NewCLIContext().
		WithCodec(app.MakeCodec()).
		WithTrustNode(true).
		WithSimulation(false).
		WithGenerateOnly(false).
		WithBroadcastMode(flags.BroadcastSync)

	clictx.SkipConfirm = true
	svc, err := exchange.NewService(&clictx, *s, ws, log.Logger)
	if nil != err {
		return errors.Wrapf(err, "fail to create service")
	}
	if err := svc.Start(); nil != err {
		log.Error().Err(err).Msg("fail to start")
	}
	//if err := svc.Stake("johnny", "BNB", "1", "1", clictx.GetFromAddress(),"welcome@1"); nil != err {
	//	panic(err)
	//}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-ch
	log.Info().Msg("quit signal receied")
	// wait for exist
	if err := svc.Stop(); nil != err {
		log.Error().Err(err).Msg("fail to stop")
	}
	if err := ds.Close(); nil != err {
		log.Error().Err(err).Msg("fail to close datastore")
	}
	return nil
}

// initLog setup logging
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
