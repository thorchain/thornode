package txsigner

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/pkg/blockclients"
	"gitlab.com/thorchain/thornode/bifrostv2/thorclient"
)

// TxSigner represents a transaction signer
type TxSigner struct {
	cfg        config.TxSignerConfigurations
	logger     zerolog.Logger
	thorClient *thorclient.Client
	chains     []blockclients.BlockChainClient
}

// NewTxSigner instantiates TxSigner
func NewTxSigner(cfg config.TxSignerConfigurations, thorClient *thorclient.Client) (*TxSigner, error) {
	return &TxSigner{
		logger:     log.Logger.With().Str("module", "txSigner").Logger(),
		cfg:        cfg,
		thorClient: thorClient,
		chains:     blockclients.LoadChains(cfg.BlockChains),
	}, nil
}

// Start starts TxSigner, listening new block on ThorChain
func (s *TxSigner) Start() error {
	return nil
}

// Stop stops TxSigner, close connections and channels
func (s *TxSigner) Stop() error {
	return nil
}
