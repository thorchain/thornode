package bifrost

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/metrics"
	"gitlab.com/thorchain/thornode/bifrostv2/thorclient"
	"gitlab.com/thorchain/thornode/bifrostv2/txscanner"
	"gitlab.com/thorchain/thornode/bifrostv2/txsigner"
	"gitlab.com/thorchain/thornode/bifrostv2/vaultmanager"
)

type Bifrost struct {
	cfg          config.Configuration
	logger       zerolog.Logger
	thorClient   *thorclient.Client
	metrics      *metrics.Metrics
	errCounter   *prometheus.CounterVec
	txScanner    *txscanner.TxScanner
	txSigner     *txsigner.TxSigner
	vaultManager *vaultmanager.VaultManager
}

func NewBifrost(cfg config.Configuration) (*Bifrost, error) {
	metric, err := metrics.NewMetrics(cfg.Metric)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create metric instance")
	}

	thorClient, err := thorclient.NewClient(cfg.ThorChain, metric)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create thorChain bridge")
	}

	vaultMgr, err := vaultmanager.NewVaultManager(cfg.ThorChain, metric)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vault manager")
	}

	txScanner := txscanner.NewTxScanner(cfg.TxScanner, vaultMgr, thorClient)

	txSigner, err := txsigner.NewTxSigner()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create txSigner")
	}

	return &Bifrost{
		cfg:          cfg,
		logger:       log.Logger.With().Str("module", "biFrost").Logger(),
		thorClient:   thorClient,
		metrics:      metric,
		txScanner:    txScanner,
		txSigner:     txSigner,
		errCounter:   metric.GetCounterVec(metrics.ObserverError),
		vaultManager: vaultMgr,
	}, nil
}

// Start, started the bifrost server and all its components
func (b *Bifrost) Start() error {
	// if err := b.metrics.Start(); err != nil {
	// 	b.logger.Error().Err(err).Msg("fail to start metric collector")
	// 	return errors.Wrap(err, "fail to start metric collector")
	// }

	// if err := b.thorClient.Start(); err != nil {
	// 	b.logger.Error().Err(err).Msg("fail to start thorchain bridge")
	// 	return errors.Wrap(err, "fail to start thorchain bridge")
	// }

	if err := b.txScanner.Start(); err != nil {
		b.logger.Error().Err(err).Msg("fail to start txScanner")
		return errors.Wrap(err, "fail to start txScanner")
	}

	if err := b.txSigner.Start(); err != nil {
		b.logger.Error().Err(err).Msg("fail to start txSigner")
		return errors.Wrap(err, "fail to start txSigner")
	}
	return nil
}

// Stop, stops the bifrost server and all its componets
func (b *Bifrost) Stop() error {
	b.logger.Info().Msg("requested to stop bifrost")
	defer b.logger.Info().Msg("bifrost stopped")

	if err := b.txScanner.Stop(); err != nil {
		b.logger.Error().Err(err).Msg("fail to stop txScanner")
	}

	if err := b.txSigner.Stop(); err != nil {
		b.logger.Error().Err(err).Msg("fail to stop txSigner")
	}

	if err := b.vaultManager.Stop(); err != nil {
		b.logger.Error().Err(err).Msg("fail to stop address manager")
	}

	if err := b.metrics.Stop(); err != nil {
		b.logger.Error().Err(err).Msg("failed to stop metrics")
	}

	return nil
}
