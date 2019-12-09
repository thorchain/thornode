package vaultmanager

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrostv2/metrics"
	"gitlab.com/thorchain/thornode/cmd"
	"gitlab.com/thorchain/thornode/common"
)

const (
	BaseEndpoint   = "/thorchain"
	VaultsEndpoint = "/vaults/pubkeys"
)

type VaultManager struct {
	logger    zerolog.Logger
	client    *retryablehttp.Client
	m         *metrics.Metrics
	chainHost string
}

func NewVaultManager(chainHost string, m *metrics.Metrics) (*VaultManager, error) {
	// Set default address config
	cosmosSDKConfg := sdk.GetConfig()
	cosmosSDKConfg.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	cosmosSDKConfg.Seal()

	return &VaultManager{
		logger:    log.With().Str("module", "VaultManager").Logger(),
		client:    retryablehttp.NewClient(),
		m:         m,
		chainHost: chainHost,
	}, nil
}

func (vaultMgr *VaultManager) Stop() error {
	return nil
}

// TODO replace to thorNode's code once endpoint is build.
type Vaults struct {
	Asgard    []common.PubKey `json:"asgard"`
	Yggdrasil []common.PubKey `json:"yggdrasil"`
}

func (vaultMgr *VaultManager) getVaults() (Vaults, error) {
	resp, err := vaultMgr.client.Get(vaultMgr.chainHost + BaseEndpoint + VaultsEndpoint)
	if err != nil {
		return Vaults{}, errors.Wrap(err, "fail to get from thorchain")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			vaultMgr.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return Vaults{}, errors.New("fail to get last block height from thorchain")

	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Vaults{}, errors.Wrap(err, "failed to read response body")
	}
	var vaults Vaults
	if err := json.Unmarshal(buf, &vaults); err != nil {
		return Vaults{}, errors.Wrap(err, "failed to unmarshal vaults")
	}
	return vaults, nil
}
