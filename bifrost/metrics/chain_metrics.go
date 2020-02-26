package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"gitlab.com/thorchain/thornode/common"
)

func SignAndBroadcastDuration(chain common.Chain) MetricName {
	return MetricName(chain + "_sign_and_broadcast_duration")
}

func TxSigned(chain common.Chain) MetricName {
	return MetricName(chain + "_tx_signed")
}

func TxSignedBroadcast(chain common.Chain) MetricName {
	return MetricName(chain + "_tx_signed_broadcast")
}

func BlockScanError(chain common.Chain) MetricName {
	return MetricName(chain + "_block_scan_error")
}

func BlockWithoutTx(chain common.Chain) MetricName {
	return MetricName(chain + "_block_without_tx")
}

func BlockWithTxIn(chain common.Chain) MetricName {
	return MetricName(chain + "_block_with_tx_in")
}

func BlockNoTxIn(chain common.Chain) MetricName {
	return MetricName(chain + "_block_no_tx_in")
}

func BlockWithTxOut(chain common.Chain) MetricName {
	return MetricName(chain + "_block_with_tx_out")
}

func BlockNoTxOut(chain common.Chain) MetricName {
	return MetricName(chain + "_block_no_tx_out")
}

func SearchTxDuration(chain common.Chain) MetricName {
	return MetricName(chain + "_search_tx_duration")
}

func AddChainMetrics(chain common.Chain, counters map[MetricName]prometheus.Counter, counterVecs map[MetricName]*prometheus.CounterVec, histograms map[MetricName]prometheus.Histogram) {
	counters[BlockWithoutTx(chain)] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "block_scanner",
		Subsystem: chain.String() + "_block_scanner",
		Name:      chain.String() + "_block_without_tx",
		Help:      "block that has no tx in it",
	})
	counters[BlockWithTxIn(chain)] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "block_scanner",
		Subsystem: chain.String() + "_block_scanner",
		Name:      chain.String() + "_block_with_tx_in",
		Help:      "block that has tx we need to process",
	})
	counters[BlockNoTxIn(chain)] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "block_scanner",
		Subsystem: chain.String() + "_block_scanner",
		Name:      chain.String() + "_block_no_tx_in",
		Help:      "block that has tx , but not to our pool address",
	})
	counters[BlockNoTxOut(chain)] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "block_scanner",
		Subsystem: chain.String() + "_block_scanner",
		Name:      chain.String() + "_block_no_tx_out",
		Help:      "block doesn't have any tx out",
	})
	counters[TxSignedBroadcast(chain)] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "signer",
		Subsystem: chain.String(),
		Name:      chain.String() + "_tx_signed_broadcast",
		Help:      "number of tx observer broadcast to " + chain.String() + " successfully",
	})
	counters[TxSigned(chain)] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "signer",
		Subsystem: chain.String(),
		Name:      chain.String() + "_tx_signed",
		Help:      "number of tx signer signed successfully",
	})

	counterVecs[BlockScanError(chain)] = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "block_scanner",
		Subsystem: chain.String() + "_block_scanner",
		Name:      "errors",
		Help:      "errors in " + chain.String() + " block scanner",
	}, []string{
		"error_name", "additional",
	})

	histograms[SearchTxDuration(chain)] = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "block_scanner",
		Subsystem: chain.String() + "_block_scanner",
		Name:      chain.String() + "_search_tx_duration",
		Help:      "how long it takes to search tx in a block in " + chain.String(),
	})
	histograms[SignAndBroadcastDuration(chain)] = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "signer",
		Subsystem: chain.String(),
		Name:      chain.String() + "_sign_and_broadcast_duration",
		Help:      "how long it takes to sign and broadcast to " + chain.String(),
	})
}
