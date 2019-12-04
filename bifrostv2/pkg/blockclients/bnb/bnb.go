package bnb

import (
	"errors"
	"strings"

	"github.com/binance-chain/go-sdk/client/rpc"
	"github.com/binance-chain/go-sdk/common/types"
	"github.com/openlyinc/pointy"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
)

type Client struct {
	client                 *rpc.HTTP
	cfg                    config.BNBConfiguration
	lastScannedBlockHeight uint64
}

func NewClient(cfg config.BNBConfiguration, lastScannedBlockHeight uint64) (*Client, error) {
	network := setNetwork(cfg)
	if cfg.ChainHost == "" {
		return nil, errors.New("chain_host not set")
	}

	return &Client{
		cfg:                    cfg,
		client:                 rpc.NewRPCClient(cfg.ChainHost, network),
		lastScannedBlockHeight: lastScannedBlockHeight,
	}, nil
}

func setNetwork(cfg config.BNBConfiguration) types.ChainNetwork {
	var network types.ChainNetwork
	if cfg.ChainNetwork == strings.ToLower("mainnet") {
		network = types.ProdNetwork
	}

	if cfg.ChainNetwork == strings.ToLower("testnet") || cfg.ChainNetwork == "" {
		network = types.TestNetwork
	}
	return network
}

func (c *Client) getBlock(blockHeight int64) (*ctypes.ResultBlock, error) {
	return c.client.Block(pointy.Int64(blockHeight))
}

func (c *Client) Start() error {
	return nil
}

func (c *Client) Stop() error {
	return nil
}
