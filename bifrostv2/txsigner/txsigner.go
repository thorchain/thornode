package txsigner

import (
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/pkg/blockclients"
	"gitlab.com/thorchain/thornode/bifrostv2/thorclient"
	"gitlab.com/thorchain/thornode/bifrostv2/types"
)

// TxSigner represents a transaction signer
type TxSigner struct {
	cfg         config.TxSignerConfigurations
	logger      zerolog.Logger
	thorClient  *thorclient.Client
	chains      []blockclients.BlockChainClient
	blockInChan chan types.Block
	wg          sync.WaitGroup
}

// NewTxSigner instantiates TxSigner
func NewTxSigner(cfg config.TxSignerConfigurations, thorClient *thorclient.Client) (*TxSigner, error) {
	return &TxSigner{
		logger:      log.Logger.With().Str("module", "txSigner").Logger(),
		cfg:         cfg,
		thorClient:  thorClient,
		chains:      blockclients.LoadChains(cfg.BlockChains),
		blockInChan: make(chan types.Block),
		wg:          sync.WaitGroup{},
	}, nil
}

// Start starts TxSigner, listening new block on ThorChain
func (s *TxSigner) Start() error {
	err := s.thorClient.Start()
	if err != nil {
		return err
	}
	s.wg.Add(1)
	go s.processBlocks(s.blockInChan)
	return nil
}

// Stop stops TxSigner, close connections and channels
func (s *TxSigner) Stop() error {
	return nil
}

func (s *TxSigner) processBlocks(blockInChan chan types.Block) {

}
