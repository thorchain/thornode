package bnb

import (
	"strings"
	"time"

	"github.com/binance-chain/go-sdk/client/rpc"
	btypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/cenkalti/backoff"
	"github.com/openlyinc/pointy"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/txscanner/types"
	"gitlab.com/thorchain/thornode/common"
)

type Client struct {
	client                   *rpc.HTTP
	cfg                      config.ChainConfigurations
	lastScannedBlockHeight   uint64
	fnLastScannedBlockHeight types.FnLastScannedBlockHeight
	logger                   zerolog.Logger
	backOffCtrl              backoff.ExponentialBackOff
}

func NewClient(cfg config.ChainConfigurations) (*Client, error) {
	network := setNetwork(cfg)
	if cfg.ChainHost == "" {
		return nil, errors.New("chain_host not set")
	}

	return &Client{
		cfg:    cfg,
		client: rpc.NewRPCClient(cfg.ChainHost, network),
		logger: log.Logger.With().Str("module", "bnbClient").Logger(),
		backOffCtrl: backoff.ExponentialBackOff{
			InitialInterval:     cfg.BackOff.InitialInterval,
			RandomizationFactor: cfg.BackOff.RandomizationFactor,
			Multiplier:          cfg.BackOff.Multiplier,
			MaxInterval:         cfg.BackOff.MaxInterval,
			MaxElapsedTime:      cfg.BackOff.MaxElapsedTime,
			Clock:               backoff.SystemClock,
		},
	}, nil
}

func setNetwork(cfg config.ChainConfigurations) btypes.ChainNetwork {
	var network btypes.ChainNetwork
	if cfg.ChainNetwork == strings.ToLower("mainnet") {
		network = btypes.ProdNetwork
	}

	if cfg.ChainNetwork == strings.ToLower("testnet") || cfg.ChainNetwork == "" {
		network = btypes.TestNetwork
	}
	return network
}

func (c *Client) getBlock(blockHeight uint64) (*ctypes.ResultBlock, error) {
	return c.client.Block(pointy.Int64(int64(blockHeight)))
}

func (c *Client) Start(txInChan chan<- types.TxIn, fnStartHeight types.FnLastScannedBlockHeight) error {
	c.logger.Info().Msg("starting")
	c.fnLastScannedBlockHeight = fnStartHeight
	c.backOffCtrl.Reset() // Reset/set the backOffCtrl

	var err error
	c.lastScannedBlockHeight, err = c.fnLastScannedBlockHeight(common.BNBChain)
	if err != nil {
		return errors.Wrap(err, "fnLastScannedBlockHeight")
	}

	go c.scanBlocks(txInChan)

	return nil
}

func (c *Client) Stop() error {
	c.logger.Info().Msg("stopped")
	return nil
}

func (c *Client) scanBlocks(txInChan chan<- types.TxIn) {
	c.logger.Info().Msg("scanBlocks")
	for {
		block, err := c.getBlock(c.lastScannedBlockHeight)
		if err != nil {
			d := c.backOffCtrl.NextBackOff()
			c.logger.Error().Err(err).Uint64("lastScannedBlockHeight", c.lastScannedBlockHeight).Str("backOffCtrl", d.String()).Msg("getBlock failed")
			time.Sleep(d)
			continue
		}

		// extract TxIns from block
		var txIn types.TxIn
		txIn.BlockHeight = uint64(block.Block.Header.Height)
		txIn.BlockHash = block.Block.Hash().String()
		txIn.Chain = common.BNBChain

		txInChan <- txIn
		c.lastScannedBlockHeight++

		c.backOffCtrl.Reset()
	}
}
