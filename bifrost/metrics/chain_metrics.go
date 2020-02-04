package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func SignAndBroadcastDuration(chain string) MetricName {
	return MetricName(chain + "_sign_and_broadcast_duration")
}

func TxSigned(chain string) MetricName {
	return MetricName(chain + "_tx_signed")
}

func TxSignedBroadcast(chain string) MetricName {
	return MetricName(chain + "_tx_signed_broadcast")
}

func BlockScanError(chain string) MetricName {
	return MetricName(chain + "_block_scan_error")
}

func BlockWithoutTx(chain string) MetricName {
	return MetricName(chain + "_block_without_tx")
}

func BlockWithTxIn(chain string) MetricName {
	return MetricName(chain + "_block_with_tx_in")
}

func BlockNoTxIn(chain string) MetricName {
	return MetricName(chain + "_block_no_tx_in")
}

func BlockWithTxOut(chain string) MetricName {
	return MetricName(chain + "_block_with_tx_out")
}

func BlockNoTxOut(chain string) MetricName {
	return MetricName(chain + "_block_no_tx_out")
}

func SearchTxDuration(chain string) MetricName {
	return MetricName(chain + "_search_tx_duration")
}

func AddChainMetrics(chain string, counters map[MetricName]prometheus.Counter, counterVecs map[MetricName]*prometheus.CounterVec, histograms map[MetricName]prometheus.Histogram) {
	counters[BlockWithoutTx(chain)] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "block_scanner",
		Subsystem: chain + "_block_scanner",
		Name:      chain + "_block_without_tx",
		Help:      "block that has no tx in it",
	})
	counters[BlockWithTxIn(chain)] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "block_scanner",
		Subsystem: chain + "_block_scanner",
		Name:      chain + "_block_with_tx_in",
		Help:      "block that has tx we need to process",
	})
	counters[BlockNoTxIn(chain)] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "block_scanner",
		Subsystem: chain + "_block_scanner",
		Name:      chain + "_block_no_tx_in",
		Help:      "block that has tx , but not to our pool address",
	})
	counters[BlockNoTxOut(chain)] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "block_scanner",
		Subsystem: chain + "_block_scanner",
		Name:      chain + "_block_no_tx_out",
		Help:      "block doesn't have any tx out",
	})
	counters[TxSignedBroadcast(chain)] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "signer",
		Subsystem: chain,
		Name:      chain + "_tx_signed_broadcast",
		Help:      "number of tx observer broadcast to " + chain + " successfully",
	})
	counters[TxSigned(chain)] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "signer",
		Subsystem: chain,
		Name:      chain + "_tx_signed",
		Help:      "number of tx signer signed successfully",
	})

	counterVecs[BlockScanError(chain)] = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "block_scanner",
		Subsystem: chain + "_block_scanner",
		Name:      "errors",
		Help:      "errors in " + chain + " block scanner",
	}, []string{
		"error_name", "additional",
	})

	histograms[SearchTxDuration(chain)] = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "block_scanner",
		Subsystem: chain + "_block_scanner",
		Name:      chain + "_search_tx_duration",
		Help:      "how long it takes to search tx in a block in " + chain,
	})
	histograms[SignAndBroadcastDuration(chain)] = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "signer",
		Subsystem: chain,
		Name:      chain + "_sign_and_broadcast_duration",
		Help:      "how long it takes to sign and broadcast to " + chain,
	})
}
