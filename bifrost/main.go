package bifrost

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/bifrost/vaultmanager"
	"gitlab.com/thorchain/thornode/cmd"
)

type Bifrost struct {
	cfg             config.Configuration
	logger          zerolog.Logger
	thorchainClient *thorclient.ThorchainBridge
	metrics         *metrics.Metrics
	errCounter      *prometheus.CounterVec
	vaultManager    *vaultmanager.VaultManager
}

func NewBifrost(cfg config.Configuration) (*Bifrost, error) {
	metric, err := metrics.NewMetrics(cfg.Metric)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create metric instance")
	}

	thorchainClient, err := thorclient.NewThorchainBridge(cfg.Thorchain, metric)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create thorchain bridge")
	}

	vaultManager, err := vaultmanager.NewVaultManager(thorchainClient, metric)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault manager")
	}

	return &Bifrost{
		cfg:             cfg,
		logger:          log.Logger.With().Str("module", "biFrost").Logger(),
		thorchainClient: thorchainClient,
		metrics:         metric,
		errCounter:      metric.GetCounterVec(metrics.ObserverError),
		vaultManager:    vaultManager,
	}, nil
}

// Start starts the bifrost server and all its components
func (b *Bifrost) Start() error {

	initPrefix()

	if err := b.metrics.Start(); err != nil {
		b.logger.Error().Err(err).Msg("fail to start metric collector")
		return errors.Wrap(err, "fail to start metric collector")
	}

	if err := b.vaultManager.Start(); err != nil {
		b.logger.Error().Err(err).Msg("fail to start vault manager")
		return errors.Wrap(err, "fail to start vault manager")
	}

	return nil
}

// Stop stops the bifrost server and all its componets
func (b *Bifrost) Stop() error {
	b.logger.Info().Msg("requested to stop bifrost")
	defer b.logger.Info().Msg("bifrost stopped")

	if err := b.vaultManager.Stop(); err != nil {
		b.logger.Error().Err(err).Msg("fail to stop address manager")
	}

	if err := b.metrics.Stop(); err != nil {
		b.logger.Error().Err(err).Msg("failed to stop metrics")
	}

	return nil
}

func initPrefix() {
	cosmosSDKConfg := sdk.GetConfig()
	cosmosSDKConfg.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	cosmosSDKConfg.Seal()
}
