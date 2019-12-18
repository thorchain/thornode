package eth

import (
	"context"
	"math/big"
	"time"

	"github.com/cenkalti/backoff"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrostv2/txblockscanner/types"
	"gitlab.com/thorchain/thornode/common"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
)

type Client struct {
	client                   *ethclient.Client
	cfg                      config.ChainConfigurations
	ctx                      context.Context
	logger                   zerolog.Logger
	fnLastScannedBlockHeight types.FnLastScannedBlockHeight
	lastScannedBlockHeight   uint64
	backOffCtrl              backoff.ExponentialBackOff
}

func NewClient(cfg config.ChainConfigurations) (*Client, error) {
	ctx := context.Background()
	ethClient, err := ethclient.DialContext(ctx, cfg.ChainHost)
	if err != nil {
		return nil, err
	}

	return &Client{
		logger: log.Logger.With().Str("module", "ethClient").Logger(),
		cfg:    cfg,
		client: ethClient,
		ctx:    ctx,
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

func (c *Client) getBlock(blockNumber uint64) (*etypes.Block, error) {
	return c.client.BlockByNumber(c.ctx, big.NewInt(int64(blockNumber)))
}

func (c *Client) getCurrentBlock() (*etypes.Block, error) {
	return c.client.BlockByNumber(c.ctx, nil)
}

func (c *Client) Start(blockInChan chan<- types.Block, fnStartHeight types.FnLastScannedBlockHeight) error {
	c.logger.Info().Msg("starting")
	c.fnLastScannedBlockHeight = fnStartHeight
	c.backOffCtrl.Reset() // Reset/set the backOffCtrl

	var err error
	c.lastScannedBlockHeight, err = c.fnLastScannedBlockHeight(common.ETHChain)
	if err != nil {
		return errors.Wrap(err, "fnLastScannedBlockHeight")

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
			c.logger.Error().Err(err).Uint64("lastScannedBlockHeight", c.lastScannedBlockHeight).Str("backoffCtrl", d.String()).Msg("getBlock failed")
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

func (c *Client) processBlock(block *etypes.Block) types.Block {
	var b types.Block
	b.BlockHeight = block.Number().Uint64()
	b.BlockHash = block.Hash().String()
	b.Chain = common.ETHChain
	// TODO extract tx data
	return b
}
