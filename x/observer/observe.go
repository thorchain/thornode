package observer

import (
	"encoding/json"

	log "github.com/rs/zerolog/log"

	"gitlab.com/thorchain/bepswap/observe/x/statechain"
	"gitlab.com/thorchain/bepswap/observe/common/types"
)

type Observer struct {
	Socket *Socket
	Scanner *Scanner
	TxChan chan []byte
}

func NewObserver(txChan chan []byte) *Observer {
	socket := NewSocket()
	scanner := NewScanner()

	return &Observer{
		Socket: socket,
		Scanner: scanner,
		TxChan: txChan,
	}
}

func (o Observer) Start() {
	sockChan := make(chan []byte)
	scanChan := make(chan []byte)

	go o.Socket.Start(sockChan)
	go o.ProcessTxn(sockChan, scanChan)
}

func (o Observer) ProcessTxn(sockChan, scanChan chan []byte) {
	for {
		var inTx types.InTx
		payload := <-sockChan

		err := json.Unmarshal(payload, &inTx)
		if err != nil {
			log.Error().Msgf("Error: %v", err)
		}

		log.Info().Msgf("Processing Transaction: %v", inTx)
		signed := statechain.Sign(inTx)
		go statechain.Send(signed, o.TxChan)

		var blocks []int
		blocks = append(blocks, inTx.BlockHeight)

		go o.Send(scanChan)
		//go o.Scanner.Scan(blocks, scanChan)
	}
}

func (o Observer) Send(scanChan chan []byte) {
	for {
		var inTx types.InTx
		payload := <-scanChan

		err := json.Unmarshal(payload, &inTx)
		if err != nil {
			log.Error().Msgf("Error: %v", err)
		}

		log.Info().Msgf("Processing Transaction: %v", inTx)

		signed := statechain.Sign(inTx)
		go statechain.Send(signed, o.TxChan)
	}
}
