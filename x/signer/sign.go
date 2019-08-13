package signer

import (
	"time"

	log "github.com/rs/zerolog/log"
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
		time.Sleep(2*time.Second)

		blockHeight := s.StateChain.TxnBlockHeight(string(txn))
		log.Info().Msgf("Got a block height of %v from StateChain", blockHeight)

		txOut := s.StateChain.TxOut(blockHeight)
		log.Info().Msgf("Got a TxOut array of %v from StateChain", txOut)

		hexTx, param := s.Binance.SignTx(txOut)
		log.Info().Msgf("Signature generated: %v", hexTx)

		s.Binance.BroadcastTx(hexTx, param)
	}
}
