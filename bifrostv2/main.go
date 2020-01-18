package bifrost

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/metrics"
	"gitlab.com/thorchain/thornode/bifrostv2/thorchain"
	"gitlab.com/thorchain/thornode/bifrostv2/txblockscanner"
	"gitlab.com/thorchain/thornode/bifrostv2/txsigner"
	"gitlab.com/thorchain/thornode/bifrostv2/vaultmanager"
	"gitlab.com/thorchain/thornode/cmd"
)

type Bifrost struct {
	cfg             config.Configuration
	logger          zerolog.Logger
	thorchainClient *thorchain.Client
	metrics         *metrics.Metrics
	errCounter      *prometheus.CounterVec
	txScanner       *txblockscanner.TxBlockScanner
	txSigner        *txsigner.TxSigner
	vaultManager    *vaultmanager.VaultManager
}

func NewBifrost(cfg config.Configuration) (*Bifrost, error) {
	metric, err := metrics.NewMetrics(cfg.Metric)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create metric instance")
	}

	thorchainClient, err := thorchain.NewClient(cfg.Thorchain, metric)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create thorChain bridge")
	}

	vaultMgr, err := vaultmanager.NewVaultManager(thorchainClient, metric)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault manager")
	}

	txScanner := txblockscanner.NewTxBlockScanner(cfg.TxScanner, vaultMgr, thorchainClient)

	txSigner, err := txsigner.NewTxSigner(cfg.TxSigner, vaultMgr, thorchainClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create txSigner")
	}

	return &Bifrost{
		cfg:             cfg,
		logger:          log.Logger.With().Str("module", "biFrost").Logger(),
		thorchainClient: thorchainClient,
		metrics:         metric,
		txScanner:       txScanner,
		txSigner:        txSigner,
		errCounter:      metric.GetCounterVec(metrics.ObserverError),
		vaultManager:    vaultMgr,
	}, nil
}

// Start starts the bifrost server and all its components
func (b *Bifrost) Start() error {

	initPrefix()

	if err := b.metrics.Start(); err != nil {
		b.logger.Error().Err(err).Msg("fail to start metric collector")
		return errors.Wrap(err, "fail to start metric collector")
	}

	if err := b.thorchainClient.Start(); err != nil {
		b.logger.Error().Err(err).Msg("fail to start thorchain client")
		return errors.Wrap(err, "fail to start thorchain client")
	}

	if err := b.vaultManager.Start(); err != nil {
		b.logger.Error().Err(err).Msg("fail to start vault manager")
		return errors.Wrap(err, "fail to start vault manager")
	}

	if err := b.txScanner.Start(); err != nil {
		b.logger.Error().Err(err).Msg("fail to start txscanner")
		return errors.Wrap(err, "fail to start txscanner")
	}

	if err := b.txSigner.Start(); err != nil {
		b.logger.Error().Err(err).Msg("fail to start txsigner")
		return errors.Wrap(err, "fail to start txsigner")
	}
	return nil
}

// Stop stops the bifrost server and all its componets
func (b *Bifrost) Stop() error {
	b.logger.Info().Msg("requested to stop bifrost")
	defer b.logger.Info().Msg("bifrost stopped")

	if err := b.txScanner.Stop(); err != nil {
		b.logger.Error().Err(err).Msg("fail to stop txscanner")
	}

	if err := b.txSigner.Stop(); err != nil {
		b.logger.Error().Err(err).Msg("fail to stop txsigner")
	}

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
