package observer

import (
	"encoding/json"

	log "github.com/rs/zerolog/log"

	types "gitlab.com/thorchain/bepswap/observe/common/types"
)

type Observer struct {
	Socket *Socket
	Scanner *Scanner
	StateChain *StateChain
}

func NewObserver(poolAddress, dexHost, rpcHost, chainHost, runeAddress string, txChan chan []byte) *Observer {
	socket := NewSocket(poolAddress, dexHost)
	scanner := NewScanner(poolAddress, dexHost, rpcHost)
	stateChain := NewStateChain(chainHost, runeAddress, txChan)

	return &Observer{
		Socket: socket,
		Scanner: scanner,
		StateChain: stateChain,
	}
}

func (o *Observer) Start() {
	sockChan := make(chan []byte)
	scanChan := make(chan []byte)

	go o.Socket.Start(sockChan)
	go o.ProcessTxn(sockChan, scanChan)
}

func (o *Observer) ProcessTxn(sockChan, scanChan chan []byte) {
	for {
		var inTx types.InTx
		payload := <-sockChan

		err := json.Unmarshal(payload, &inTx)
		if err != nil {
			log.Error().Msgf("%s Error: %v", LogPrefix(), err)
		}

		log.Info().Msgf("%s Processing Transaction: %v", LogPrefix(), inTx)
		go o.StateChain.Send(inTx)

		var blocks []int
		blocks = append(blocks, inTx.BlockHeight)

		go o.Send(scanChan)
		//go o.Scanner.Scan(blocks, scanChan)
	}
}

func (o *Observer) Send(scanChan chan []byte) {
	for {
		var inTx types.InTx
		payload := <-scanChan

		err := json.Unmarshal(payload, &inTx)
		if err != nil {
			log.Error().Msgf("%s Error: %v", LogPrefix(), err)
		}

		log.Info().Msgf("%s Processing Transaction: %v", LogPrefix(), inTx)
		go o.StateChain.Send(inTx)
	}
}

func LogPrefix() string { return "[OBSERVER]" }
