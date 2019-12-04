package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"

	"gitlab.com/thorchain/thornode/bifrostv2"
	"gitlab.com/thorchain/thornode/bifrostv2/config"
)

// we define version / revision here , so we could inject the version from CI pipeline if we want to
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
	// TODO set the default log level to info later
	logLevel := flag.StringP("log-level", "l", "debug", "Log Level")
	pretty := flag.BoolP("pretty-log", "p", false, "Enables unstructured prettified logging. This is useful for local debugging")
	cfgFile := flag.StringP("cfg", "c", "config", "configuration file with extension")
	flag.Parse()
	if *showVersion {
		printVersion()
		return
	}
	initLog(*logLevel, *pretty)
	cfg, err := config.LoadBiFrostConfig(*cfgFile)
	if nil != err {
		log.Fatal().Err(err).Msg("fail to load bifrost config ")
	}

	bi, err := bifrost.NewBifrost(*cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to create bifrost")
	}

	if err := bi.Start(); nil != err {
		log.Fatal().Err(err).Msg("fail to start bifrost")
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	for {
		select {
		case <-interrupt:
			log.Info().Msg("os stop signal received")
			if err := bi.Stop(); nil != err {
				log.Fatal().Err(err).Msg("fail to stop bitfrost")
			}
			return
		}
	}
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
	if level == "debug" {
		log.Logger = log.With().Caller().Logger()
	}
	zerolog.SetGlobalLevel(l)
	log.Logger = log.Output(out).With().Str("service", serverIdentity).Logger()
}
