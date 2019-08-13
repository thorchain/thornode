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
		log.Info().Msgf("%s Received Transaction: %v", LogPrefix(), string(txn))

		blockHeight := s.StateChain.TxnBlockHeight(string(txn))
		log.Info().Msgf("%s Received a Block Height of %v from StateChain", LogPrefix(), blockHeight)

		txOut := s.StateChain.TxOut(blockHeight)
		log.Info().Msgf("%s Received a TxOut Array of %v from StateChain", LogPrefix(), txOut)

		hexTx, param := s.Binance.SignTx(txOut)
		log.Info().Msgf("%s Generated the following signature for Binance: %v", LogPrefix(), string(hexTx))

		s.Binance.BroadcastTx(hexTx, param)
	}
}

func LogPrefix() string {
	return "[SIGNER]"
}
