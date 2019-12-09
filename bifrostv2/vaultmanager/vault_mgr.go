package vaultmanager

import (
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrostv2/metrics"
	"gitlab.com/thorchain/thornode/bifrostv2/thorclient"
)

type VaultManager struct {
	logger     zerolog.Logger
	client     *retryablehttp.Client
	m          *metrics.Metrics
	thorClient *thorclient.Client
}

func NewVaultManager(thorClient *thorclient.Client, m *metrics.Metrics) (*VaultManager, error) {
	return &VaultManager{
		logger:     log.With().Str("module", "VaultManager").Logger(),
		client:     retryablehttp.NewClient(),
		m:          m,
		thorClient: thorClient,
	}, nil
}

func (vaultMgr *VaultManager) Start() error {
	return nil
}

func (vaultMgr *VaultManager) Stop() error {
	return nil
}
