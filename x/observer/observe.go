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
	ScanClient *ScanClient
	StateChain *StateChain
}

func NewObserver(poolAddress, dexHost, rpcHost, chainHost, runeAddress string, txChan chan []byte) *Observer {
	socketClient := NewSocketClient(poolAddress, dexHost)
	scanClient := NewScanClient(poolAddress, dexHost, rpcHost)
	stateChain := NewStateChain(chainHost, runeAddress, txChan)

	return &Observer{
		PoolAddress: poolAddress,
		DexHost: dexHost,
		SocketClient: socketClient,
		ScanClient: scanClient,
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
		// go c.ScanClient.ProcessBlocks(blocks, scanChan)
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
