package btc

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/txblockscanner/types"
	"gitlab.com/thorchain/thornode/common"
)

type Client struct {
	cfg                      config.ChainConfigurations
	client                   *rpcclient.Client
	logger                   zerolog.Logger
	fnLastScannedBlockHeight types.FnLastScannedBlockHeight
	lastScannedBlockHeight   uint64
	backOffCtrl              backoff.ExponentialBackOff
}

func NewClient(cfg config.ChainConfigurations) (*Client, error) {
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         cfg.ChainHost,
		User:         cfg.UserName,
		Pass:         cfg.Password,
		DisableTLS:   cfg.DisableTLS,
		HTTPPostMode: cfg.HTTPostMode,
	}, nil)
	if err != nil {
		return &Client{}, err
	}

	return &Client{
		cfg:    cfg,
		client: client,
		logger: log.Logger.With().Str("module", "btcClient").Logger(),
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

func (c *Client) Start(blockInChan chan<- types.Block, fnStartHeight types.FnLastScannedBlockHeight) error {
	c.logger.Info().Msg("starting")
	c.fnLastScannedBlockHeight = fnStartHeight
	c.backOffCtrl.Reset() // Reset/set the backOffCtrl

	var err error
	c.lastScannedBlockHeight, err = c.fnLastScannedBlockHeight(common.BTCChain)
	if err != nil {
		return errors.Wrap(err, "bitcoinClient failed")
	}

	go c.scanBlocks(blockInChan)
	return nil
}

func (c *Client) scanBlocks(blockInChan chan<- types.Block) {
	c.logger.Info().Msg("scanBlocks")
	for {
		block, err := c.getBlock(c.lastScannedBlockHeight)
		if err != nil {
			d := c.backOffCtrl.NextBackOff()
			c.logger.Error().Err(err).Uint64("lastScannedBlockHeight", c.lastScannedBlockHeight).Str("backOffCtrl", d.String()).Msg("getBlock failed")
			time.Sleep(d)
			continue
		}

		blockInChan <- c.processBlock(block)
		c.lastScannedBlockHeight++

		c.backOffCtrl.Reset()
	}
}

func (c *Client) Stop() error {
	c.logger.Info().Msg("stopped")
	return nil
}

func (c *Client) getBlock(blockHeight uint64) (*wire.MsgBlock, error) {
	hash, err := c.getBlockHash(int64(blockHeight))
	if err != nil {
		return &wire.MsgBlock{}, err
	}
	return c.client.GetBlock(hash)
}

func (c *Client) getBlockHash(blockHeight int64) (*chainhash.Hash, error) {
	hash, err := c.client.GetBlockHash(blockHeight)
	if err != nil {
		return &chainhash.Hash{}, err
	}
	return hash, nil
}

func (c *Client) processBlock(block *wire.MsgBlock) types.Block {
	var b types.Block
	b.BlockHeight = c.lastScannedBlockHeight
	b.BlockHash = block.BlockHash().String()
	b.Chain = common.BTCChain

	// TODO extract Tx data
	return b
}
