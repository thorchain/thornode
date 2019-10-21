package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
)

// Query Result Payload for a pools query
type QueryResPools []Pool

// implement fmt.Stringer
func (n QueryResPools) String() string {
	var tickers []string
	for _, record := range n {
		tickers = append(tickers, record.Ticker.String())
	}
	return strings.Join(tickers, "\n")
}

type QueryResHeights struct {
	LastBinanceHeight sdk.Uint `json:"lastobservedin"`
	LastSignedHeight  sdk.Uint `json:"lastsignedout"`
	Statechain        int64    `json:"statechain"`
}

func (h QueryResHeights) String() string {
	return fmt.Sprintf("Binance: %d, Signed: %d, Statechain: %d", h.LastBinanceHeight, h.LastSignedHeight, h.Statechain)
}

type ResTxOut struct {
	Height  uint64       `json:"height"`
	Hash    common.TxID  `json:"hash"`
	Chain   common.Chain `json:"chain"`
	TxArray []TxOutItem  `json:"tx_array"`
}

type QueryResTxOut struct {
	Chains map[common.Chain]ResTxOut `json:"chains"`
}
