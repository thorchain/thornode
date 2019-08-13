package signer

import (
	//"time"

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
		log.Info().Msgf("[SIGNER] Received Transaction: %v", string(txn))

		blockHeight := s.StateChain.TxnBlockHeight(string(txn))
		log.Info().Msgf("[SIGNER] Received a Block Height of %v from StateChain", blockHeight)

		txOut := s.StateChain.TxOut(blockHeight)
		log.Info().Msgf("[SIGNER] Received a TxOut Array of %v from StateChain", txOut)

		hexTx, param := s.Binance.SignTx(txOut)
		log.Info().Msgf("[SIGNER] Generated the following signature for Binance: %v", string(hexTx))

		s.Binance.BroadcastTx(hexTx, param)
	}
}
