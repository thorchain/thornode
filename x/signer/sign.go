package signer

import (
	log "github.com/rs/zerolog/log"

	"gitlab.com/thorchain/bepswap/observe/x/binance"
	"gitlab.com/thorchain/bepswap/observe/x/statechain"
)

type Signer struct {
	Binance *binance.Binance
	TxChan chan []byte
}

func NewSigner(txChan chan []byte) *Signer {
	binance := binance.NewBinance()

	return &Signer{
		Binance: binance,
		TxChan: txChan,
	}
}

func (s Signer) Start() {
	go s.ProcessTxn()
}

func (s Signer) ProcessTxn() {
	for {
		txn := <-s.TxChan
		log.Info().Msgf("Received a transaction hash of: %v", string(txn))

		blockHeight := statechain.TxnBlockHeight(string(txn))
		if blockHeight != "" {
			log.Info().Msgf("Received a Block Height of %v from the StateChain", blockHeight)

			txOut := statechain.TxOut(blockHeight)
			if txOut.Height != "0" && len(txOut.TxArray) >= 1 {
				log.Info().Msgf("Received a TxOut Array of %v from the StateChain", txOut)

				hexTx, param := s.Binance.SignTx(txOut)
				log.Info().Msgf("Generated a signature for Binance: %v", string(hexTx))

				s.Binance.BroadcastTx(hexTx, param)
			} else {
				log.Error().Msg("Received an empty TxOut Array from the StateChain")
			}
		} else {
			log.Error().Msg("Received an empty Block Height from the StateChain")
		}
	}
}
