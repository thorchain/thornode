package signer

import (
	log "github.com/rs/zerolog/log"

	"gitlab.com/thorchain/bepswap/observe/x/binance"
)

type Signer struct {
	Binance *binance.Binance
	StateChain *StateChain
	TxChan chan []byte
}

func NewSigner(poolAddress, dexHost, chainHost string, txChan chan []byte) *Signer {
	binance := binance.NewBinance(poolAddress, dexHost)
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
		if blockHeight != "" {
			log.Info().Msgf("%s Received a Block Height of %v from the StateChain", LogPrefix(), blockHeight)

			txOut := s.StateChain.TxOut(blockHeight)
			if txOut.Height != "0" && len(txOut.TxArray) >= 1 {
				log.Info().Msgf("%s Received a TxOut Array of %v from the StateChain", LogPrefix(), txOut)

				hexTx, param := s.Binance.SignTx(txOut)
				log.Info().Msgf("%s Generated the following signature for Binance: %v", LogPrefix(), string(hexTx))

				s.Binance.BroadcastTx(hexTx, param)
			} else {
				log.Error().Msgf("%s Received an empty TxOut Array from the StateChain", LogPrefix())
			}
		} else {
			log.Error().Msgf("%s Received an empty Block Height from the StateChain", LogPrefix())
		}
	}
}

func LogPrefix() string { return "[SIGNER]" }
