package signer

import (
	"encoding/json"

	log "github.com/rs/zerolog/log"
	"github.com/syndtr/goleveldb/leveldb"

	"gitlab.com/thorchain/bepswap/observe/x/binance"
	ctypes "gitlab.com/thorchain/bepswap/observe/common/types"
	stypes "gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

type Signer struct {
	Db *leveldb.DB
	BlockScan *BlockScan
	Binance *binance.Binance
	TxOutChan chan []byte
}

func NewSigner() *Signer {
	var db, _ = leveldb.OpenFile(ctypes.SignerDbPath, nil)

	txOutChan := make(chan []byte)
	blockScan := NewBlockScan(db, txOutChan)
	binance := binance.NewBinance()

	return &Signer{
		Db: db,
		BlockScan: blockScan,
		Binance: binance,
		TxOutChan: txOutChan,
	}
}

func (s Signer) Start() {
	go s.ProcessTxnOut()
	go s.BlockScan.Start()
}

func (s Signer) ProcessTxnOut() {
	for {
		payload := <-s.TxOutChan

		var txOut stypes.TxOut
		err := json.Unmarshal(payload, &txOut)
		if err != nil {
			log.Info().Msgf("Error: %v", err)
		}

		log.Info().Msgf("Received a TxOut Array of %v from the StateChain", txOut)

		hexTx, param := s.Binance.SignTx(txOut)
		log.Info().Msgf("Generated a signature for Binance: %v", string(hexTx))

		s.Binance.BroadcastTx(hexTx, param)
	}
}
