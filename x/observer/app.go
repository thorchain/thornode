package observer

import (
	"encoding/json"

	log "github.com/rs/zerolog/log"

	types "gitlab.com/thorchain/bepswap/observe/x/observer/types"
)

type App struct {
	PoolAddress string
	DexHost string
	SocketClient *SocketClient
	Scanner *Scanner
	StateChain *StateChain
}

func NewApp(poolAddress, dexHost, rpcHost, chainHost, runeAddress string) *App {
	socketclient := NewSocketClient(poolAddress, dexHost)
	scanner := NewScanner(poolAddress, dexHost, rpcHost)
	statechain := NewStateChain(chainHost, runeAddress)

	return &App{
		PoolAddress: poolAddress,
		DexHost: dexHost,
		SocketClient: socketclient,
		Scanner: scanner,
		StateChain: statechain,
	}
}

func (a *App) Start() {
	sockChan := make(chan []byte)
	scanChan := make(chan []byte)

	go a.SocketClient.StartClient(sockChan)
	go a.ProcessTxn(sockChan, scanChan)
}

func (a *App) ProcessTxn(sockChan, scanChan chan []byte) {
	for {
		var inTx types.InTx
		payload := <-sockChan
		log.Info().Msgf("Received Transaction: %v", string(payload))

		err := json.Unmarshal(payload, &inTx)
		if err != nil {
			log.Error().Msgf("Error: %v", err)
		}

		log.Info().Msgf("Processing Transaction: %v", inTx)
		go a.StateChain.Send(inTx)

		var blocks []int
		blocks = append(blocks, inTx.BlockHeight)

		go a.ProcessBlockTxn(scanChan)
		// go c.Scanner.ProcessBlocks(blocks, scanChan)
	}
}

func (a *App) ProcessBlockTxn(scanChan chan []byte) {
	for {
		var inTx types.InTx
		payload := <-scanChan

		err := json.Unmarshal(payload, &inTx)
		if err != nil {
			log.Error().Msgf("Error: %v", err)
		}

		log.Info().Msgf("Processing Transaction: %v", inTx)
		go a.StateChain.Send(inTx)
	}
}
