package signer

import (
	"encoding/json"

	log "github.com/rs/zerolog/log"

	types "gitlab.com/thorchain/bepswap/observe/x/signer/types"
)

type Signer struct {
	Binance *Binance
	StateChain *StateChain
	TxChan chan []byte
}

func NewSigner(poolAddress, dexHost, chainHost string, txChan chan []byte) *Signer {
	binance := NewBinance(poolAddress, dexHost)
	stateChain := NewStateChain(chainHost)

	return &Signer{
		Binance: binance,
		StateChain: stateChain,
		TxChan: txChan,
	}
}

func (s *Signer) Start() {
	go s.ProcessTxn()
}

func (s *Signer) ProcessTxn() {
	for {
		txn := <-s.TxChan
		log.Info().Msgf("Received Transaction: %v", string(txn))

		var txs types.Txs
		json.Unmarshal(txn, &txs)

		blockHeight := s.StateChain.TxnBlockHeight(txs.Height)
		txOut := s.StateChain.TxOut(blockHeight)

		hexTx, param := s.Binance.SignTx(txOut)
		s.Binance.BroadcastTx(hexTx, param)
	}
}
