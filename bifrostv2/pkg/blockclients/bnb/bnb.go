package bnb

import (
	"fmt"
	"strings"
	"time"

	"github.com/binance-chain/go-sdk/client/rpc"
	btypes "github.com/binance-chain/go-sdk/common/types"
	bmsg "github.com/binance-chain/go-sdk/types/msg"
	"github.com/cenkalti/backoff"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/openlyinc/pointy"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/txblockscanner/types"
	"gitlab.com/thorchain/thornode/common"

	"github.com/binance-chain/go-sdk/types/tx"
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

func (c *Client) Start(blockInChan chan<- types.Block, fnStartHeight types.FnLastScannedBlockHeight) error {
	c.logger.Info().Msg("starting")
	c.fnLastScannedBlockHeight = fnStartHeight
	c.backOffCtrl.Reset() // Reset/set the backOffCtrl

	var err error
	c.lastScannedBlockHeight, err = c.fnLastScannedBlockHeight(common.BNBChain)
	if err != nil {
		return errors.Wrap(err, "fnLastScannedBlockHeight")
	}

	go c.scanBlocks(blockInChan)

	return nil
}

func (c *Client) Stop() error {
	c.logger.Info().Msg("stopped")
	return nil
}

func (c *Client) scanBlocks(blockInChan chan<- types.Block) {
	c.logger.Info().Msg("scanBlocks")
	for {
		block, err := c.getBlock(c.lastScannedBlockHeight)
		if err != nil || block.Block == nil {
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

// processBlock extract's block information of interest into out generic/simplified block struct to processing by the txBlockScanner module.
func (c *Client) processBlock(block *ctypes.ResultBlock) types.Block {
	var b types.Block

	b.Chain = common.BNBChain
	b.BlockHash = block.Block.Hash().String()
	b.BlockHeight = uint64(block.Block.Header.Height)

	for _, txx := range block.Block.Data.Txs {
		var t tx.StdTx
		if err := tx.Cdc.UnmarshalBinaryLengthPrefixed(txx, &t); err != nil {
			c.logger.Err(err).Msg("UnmarshalBinaryLengthPrefixed")
		}

		txItems, err := c.processStdTx(t)
		if err != nil {
			c.logger.Err(err).Msg("failed to processStdTx")
			continue
		}

		// if valid blocks returned
		if len(txItems) > 0 {
			b.Txs = append(b.Txs, txItems...)
		}
	}
	return b
}

// processStdTx extract's tx information of interest into our generic TxItem struct
func (c *Client) processStdTx(stdTx tx.StdTx) ([]types.TxItem, error) {
	var txItems []types.TxItem
	var err error

	// TODO: it is possible to have multiple `SendMsg` in a single stdTx, which
	// THORNode are currently not accounting for. It is also possible to have
	// multiple inputs/outputs within a single stdTx, which THORNode are not yet
	// accounting for.
	for _, msg := range stdTx.Msgs {
		switch sendMsg := msg.(type) {
		case bmsg.SendMsg:
			txItem := types.TxItem{}
			txItem.Memo = stdTx.Memo

			// if no memo its not worth processing at all.
			if len(txItem.Memo) == 0 {
				continue
			}

			// THORNode take the first Input as sender, first Output as receiver
			// so if THORNode send to multiple different receiver within one tx, this won't be able to process it.
			sender := sendMsg.Inputs[0]
			receiver := sendMsg.Outputs[0]
			txItem.From = common.Address(sender.Address.String())
			txItem.To = common.Address(receiver.Address.String())

			txItem.Coins, err = c.getCoinsForTxIn(sendMsg.Outputs)
			if err != nil {
				return nil, errors.Wrap(err, "fail to convert coins")
			}

			// TODO: We should not assume what the gas fees are going to be in
			// the future (although they are largely static for binance). We
			// should modulus the binance block height and get the latest fee
			// prices every 1,000 or so blocks. This would ensure that all
			// observers will always report the same gas prices as they update
			// their price fees at the same time.

			// Calculate gas for this tx
			if len(txItem.Coins) > 1 {
				// Multisend gas fees
				txItem.Gas = common.GetBNBGasFeeMulti(uint64(len(txItem.Coins)))
			} else {
				// Single transaction gas fees
				txItem.Gas = common.BNBGasFeeSingleton
			}

			txItems = append(txItems, txItem)
		default:
			continue
		}
	}
	return txItems, nil
}

// getCoinsForTxIn extract's the coins/amount into our generic Coins struct
func (c *Client) getCoinsForTxIn(outputs []bmsg.Output) (common.Coins, error) {
	cc := common.Coins{}
	for _, output := range outputs {
		for _, c := range output.Coins {
			asset, err := common.NewAsset(fmt.Sprintf("BNB.%s", c.Denom))
			if err != nil {
				return nil, errors.Wrapf(err, "fail to create asset, %s is not valid", c.Denom)
			}
			amt := sdk.NewUint(uint64(c.Amount))
			cc = append(cc, common.NewCoin(asset, amt))
		}
	}
	return cc, nil
}
