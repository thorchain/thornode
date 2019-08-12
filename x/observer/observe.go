package observer

import (
	"encoding/json"

	log "github.com/rs/zerolog/log"

	types "gitlab.com/thorchain/bepswap/observe/x/observer/types"
)

type Observer struct {
	PoolAddress string
	DexHost string
	SocketClient *SocketClient
	Scanner *Scanner
	StateChain *StateChain
}

func NewObserver(poolAddress, dexHost, rpcHost, chainHost, runeAddress string) *Observer {
	socketClient := NewSocketClient(poolAddress, dexHost)
	scanner := NewScanner(poolAddress, dexHost, rpcHost)
	stateChain := NewStateChain(chainHost, runeAddress)

	return &Observer{
		PoolAddress: poolAddress,
		DexHost: dexHost,
		SocketClient: socketClient,
		Scanner: scanner,
		StateChain: stateChain,
	}
}

func (o *Observer) Start() {
	sockChan := make(chan []byte)
	scanChan := make(chan []byte)

	go o.SocketClient.StartClient(sockChan)
	go o.ProcessTxn(sockChan, scanChan)
}

func (o *Observer) ProcessTxn(sockChan, scanChan chan []byte) {
	for {
		var inTx types.InTx
		payload := <-sockChan
		log.Info().Msgf("Received Transaction: %v", string(payload))

		err := json.Unmarshal(payload, &inTx)
		if err != nil {
			log.Error().Msgf("Error: %v", err)
		}

		log.Info().Msgf("Processing Transaction: %v", inTx)
		go o.StateChain.Send(inTx)

		var blocks []int
		blocks = append(blocks, inTx.BlockHeight)

		go o.ProcessBlockTxn(scanChan)
		// go c.Scanner.ProcessBlocks(blocks, scanChan)
	}
}

func (o *Observer) ProcessBlockTxn(scanChan chan []byte) {
	for {
		var inTx types.InTx
		payload := <-scanChan

		err := json.Unmarshal(payload, &inTx)
		if err != nil {
			log.Error().Msgf("Error: %v", err)
		}

		log.Info().Msgf("Processing Transaction: %v", inTx)
		go o.StateChain.Send(inTx)
	}
}
