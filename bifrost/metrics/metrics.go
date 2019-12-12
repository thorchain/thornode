package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrost/config"
)

// MetricName
type MetricName string

const (
	TotalBlockScanned       MetricName = `total_block_scanned`
	CurrentPosition         MetricName = `current_position`
	TotalRetryBlocks        MetricName = `total_retry_blocks`
	CommonBLockScannerError MetricName = `block_scanner_error`

	BinanceBlockScanError MetricName = `biance_block_scan_error`
	BlockWithoutTx        MetricName = `block_no_tx`
	BlockWithTxIn         MetricName = `block_tx_in`
	BlockNoTxIn           MetricName = `block_no_tx_in`

	StateChainBlockScanError MetricName = `statechain_block_scan_error`
	BlockNoTxOut             MetricName = `block_no_txout`

	BlockDiscoveryDuration MetricName = `block_discovery_duration`
	SearchTxDuration       MetricName = `search_tx_duration`

	StateChainBridgeError    MetricName = `statechain_bridge_error`
	TxToStateChain           MetricName = `tx_to_statechain`
	TxToStateChainSigned     MetricName = `tx_to_statechain_signed`
	SignToStateChainDuration MetricName = `sign_to_statechain_duration`
	SendToStatechainDuration MetricName = `send_to_statechain_duration`

	ObserverError                     MetricName = `observer_error`
	SignerError                       MetricName = `signer_error`
	TxToBinanceSigned                 MetricName = `tx_to_binance_signed`
	TxToBinanceSignedBroadcast        MetricName = `tx_to_binance_broadcast`
	SignAndBroadcastToBinanceDuration MetricName = `sign_and_broadcast_to_binance_duration`

	PoolAddressManagerError MetricName = `pool_address_manager_error`
)

// Metrics used to provide promethus metrics
type Metrics struct {
	logger zerolog.Logger
	cfg    config.MetricConfiguration
	s      *http.Server
	wg     *sync.WaitGroup
}

var (
	counters = map[MetricName]prometheus.Counter{
		TotalBlockScanned: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "block_scanner",
			Subsystem: "common_block_scanner",
			Name:      "total_block_scanned",
			Help:      "Total number of block scanned",
		}),
		CurrentPosition: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "block_scanner",
			Subsystem: "common_block_scanner",
			Name:      "current_position",
			Help:      "current block scan position",
		}),
		TotalRetryBlocks: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "block_scanner",
			Subsystem: "common_block_scanner",
			Name:      "total_retry_blocks",
			Help:      "total blocks retried ",
		}),
		BlockWithoutTx: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "block_scanner",
			Subsystem: "binance_block_scanner",
			Name:      "block_without_tx",
			Help:      "block that has no tx in it",
		}),
		BlockWithTxIn: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "block_scanner",
			Subsystem: "binance_block_scanner",
			Name:      "block_with_tx_in",
			Help:      "block that has tx THORNode need to process",
		}),
		BlockNoTxIn: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "block_scanner",
			Subsystem: "binance_block_scanner",
			Name:      "block_no_tx_in",
			Help:      "block that has tx , but not to our pool address",
		}),
		BlockNoTxOut: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "block_scanner",
			Subsystem: "statechain_block_scanner",
			Name:      "block_no_tx_out",
			Help:      "block doesn't have any tx out",
		}),
		TxToStateChain: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "observer",
			Subsystem: "statechain_bridge",
			Name:      "tx_to_statechain",
			Help:      "number of tx observer post to statechain successfully",
		}),
		TxToStateChainSigned: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "observer",
			Subsystem: "statechain_bridge",
			Name:      "tx_to_statechain_signed",
			Help:      "number of tx observer signed successfully",
		}),
		TxToBinanceSigned: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "signer",
			Subsystem: "binance",
			Name:      "tx_to_binance_signed",
			Help:      "number of tx signer signed successfully",
		}),
		TxToBinanceSignedBroadcast: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "signer",
			Subsystem: "binance",
			Name:      "tx_to_binance_broadcast",
			Help:      "number of tx observer broadcast to binance successfully",
		}),
	}
	counterVecs = map[MetricName]*prometheus.CounterVec{
		CommonBLockScannerError: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "block_scanner",
			Subsystem: "common_block_scanner",
			Name:      "errors",
			Help:      "errors in common block scanner",
		}, []string{
			"error_name", "additional",
		}),
		BinanceBlockScanError: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "block_scanner",
			Subsystem: "binance_block_scanner",
			Name:      "errors",
			Help:      "errors in binance block scanner",
		}, []string{
			"error_name", "additional",
		}),

		StateChainBlockScanError: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "block_scanner",
			Subsystem: "statechain_block_scanner",
			Name:      "errors",
			Help:      "errors in statechain block scanner",
		}, []string{
			"error_name", "additional",
		}),

		StateChainBridgeError: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "statechain",
			Subsystem: "statechain_bridge",
			Name:      "errors",
			Help:      "errors in statechain bridge",
		}, []string{
			"error_name", "additional",
		}),

		ObserverError: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "observer",
			Subsystem: "observer",
			Name:      "errors",
			Help:      "errors in observer",
		}, []string{
			"error_name", "additional",
		}),
		SignerError: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "signer",
			Subsystem: "signer",
			Name:      "errors",
			Help:      "errors in signer",
		}, []string{
			"error_name", "additional",
		}),
		PoolAddressManagerError: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "observer",
			Subsystem: "pool_addresses_manager",
			Name:      "errors",
			Help:      "errors in pool addresses manager",
		}, []string{
			"error_name", "additional",
		}),
	}

	histograms = map[MetricName]prometheus.Histogram{
		BlockDiscoveryDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "block_scanner",
			Subsystem: "common_block_scanner",
			Name:      "block_discovery",
			Help:      "how long it takes to discovery a block height",
		}),
		SearchTxDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "block_scanner",
			Subsystem: "binance_block_scanner",
			Name:      "search_tx_duration",
			Help:      "how long it takes to search tx in a block",
		}),
		SignAndBroadcastToBinanceDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "signer",
			Subsystem: "binance",
			Name:      "sign_and_broadcast_to_binance",
			Help:      "how long it takes to sign and broadcast to binance",
		}),
		SignToStateChainDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "observer",
			Subsystem: "statechain",
			Name:      "sign_to_statechain_duration",
			Help:      "how long it takes to sign a tx to statechain",
		}),
		SendToStatechainDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "observer",
			Subsystem: "statechain",
			Name:      "send_to_statechain_duration",
			Help:      "how long it takes to sign and broadcast to binance",
		}),
	}
)

// NewMetrics create a new instance of Metrics
func NewMetrics(cfg config.MetricConfiguration) (*Metrics, error) {
	// create a new mux server
	server := http.NewServeMux()
	// register a new handler for the /metrics endpoint
	server.Handle("/metrics", promhttp.Handler())
	// start an http server using the mux server
	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ListenPort),
		Handler:      server,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	return &Metrics{
		logger: log.With().Str("module", "metrics").Logger(),
		cfg:    cfg,
		s:      s,
		wg:     &sync.WaitGroup{},
	}, nil
}

// GetCounter return a counter by name, if it doesn't exist, then it return nil
func (m *Metrics) GetCounter(name MetricName) prometheus.Counter {
	if counter, ok := counters[name]; ok {
		return counter
	}
	return nil
}

// GetHistograms return a histogram by name
func (m *Metrics) GetHistograms(name MetricName) prometheus.Histogram {
	if h, ok := histograms[name]; ok {
		return h
	}
	return nil
}

func (m *Metrics) GetCounterVec(name MetricName) *prometheus.CounterVec {
	if c, ok := counterVecs[name]; ok {
		return c
	}
	return nil
}

// Start
func (m *Metrics) Start() error {
	if !m.cfg.Enabled {
		return nil
	}
	m.wg.Add(1)
	go func() {
		m.logger.Info().Int("port", m.cfg.ListenPort).Msg("start metric server")
		if err := m.s.ListenAndServe(); nil != err {
			m.logger.Error().Err(err).Msg("fail to stop metric server")
		}
	}()
	return nil
}

// Stop
func (m *Metrics) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	return m.s.Shutdown(ctx)
}

func init() {
	for _, item := range counterVecs {
		prometheus.MustRegister(item)
	}
	for _, item := range counters {
		prometheus.MustRegister(item)
	}
	for _, item := range histograms {
		prometheus.MustRegister(item)
	}
}
