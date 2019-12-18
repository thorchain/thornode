package thorclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrostv2/keys"
	"gitlab.com/thorchain/thornode/cmd"
	"gitlab.com/thorchain/thornode/common"
	stypes "gitlab.com/thorchain/thornode/x/thorchain/types"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/metrics"
	"gitlab.com/thorchain/thornode/bifrostv2/thorclient/types"
)

const (
	BaseEndpoint   = "/thorchain"
	VaultsEndpoint = "/vaults/pubkeys"
)

// Client is for all communication to a thorNode
type Client struct {
	logger        zerolog.Logger
	cdc           *codec.Codec
	cfg           config.ThorChainConfiguration
	keys          *keys.Keys
	errCounter    *prometheus.CounterVec
	m             *metrics.Metrics
	accountNumber uint64
	seqNumber     uint64
	client        *retryablehttp.Client
}

// NewClient create a new instance of Client
func NewClient(cfg config.ThorChainConfiguration, m *metrics.Metrics) (*Client, error) {
	if len(cfg.ChainID) == 0 {
		return nil, errors.New("chain id is empty")
	}
	if len(cfg.ChainHost) == 0 {
		return nil, errors.New("chain host is empty")
	}
	if len(cfg.SignerName) == 0 {
		return nil, errors.New("signer name is empty")
	}
	if len(cfg.SignerPasswd) == 0 {
		return nil, errors.New("signer password is empty")
	}
	k, err := keys.NewKeys(cfg.ChainHomeFolder, cfg.SignerName, cfg.SignerPasswd)
	if nil != err {
		return nil, fmt.Errorf("fail to get keybase: %w", err)
	}

	return &Client{
		logger:     log.With().Str("module", "thorClient").Logger(),
		cdc:        MakeCodec(),
		cfg:        cfg,
		keys:       k,
		errCounter: m.GetCounterVec(metrics.ThorChainClientError),
		client:     retryablehttp.NewClient(), // TODO Setup a logger function that is in our format
		m:          m,
	}, nil
}

func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	sdk.RegisterCodec(cdc)
	stypes.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

// CosmosSDKConfig set's the default address prefixes from thorChain
func CosmosSDKConfig() {
	cosmosSDKConfig := sdk.GetConfig()
	cosmosSDKConfig.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	cosmosSDKConfig.Seal()
}

// Start ensure that the bifrost has been whitelisted and is ready to run.
func (c *Client) Start() error {
	CosmosSDKConfig()

	if err := c.ensureNodeWhitelistedWithTimeout(); err != nil {
		c.logger.Error().Err(err).Msg("node account is not whitelisted, can't start")
		return errors.Wrap(err, "node account is not whitelisted, can't start")
	}

	accountNumber, sequenceNumber, err := c.getAccountNumberAndSequenceNumber()
	if nil != err {
		return errors.Wrap(err, "fail to get account number and sequence number from thorchain")
	}

	c.logger.Info().Uint64("account number", accountNumber).Uint64("sequence no", sequenceNumber).Msg("account information")
	c.accountNumber = accountNumber
	c.seqNumber = sequenceNumber
	return nil
}

// getAccountNumberAndSequenceNumber returns account and Sequence number required to post into thorchain
func (c *Client) getAccountNumberAndSequenceNumber() (uint64, uint64, error) {
	requestUrl := fmt.Sprintf("/auth/accounts/%s", c.keys.GetSignerInfo().GetAddress())

	body, err := c.get(requestUrl)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to call: "+requestUrl)
	}
	var accountResp types.AccountResp
	if err := json.Unmarshal(body, &accountResp); nil != err {
		return 0, 0, errors.Wrap(err, "fail to unmarshal account resp")
	}
	var baseAccount authtypes.BaseAccount
	err = authtypes.ModuleCdc.UnmarshalJSON(accountResp.Result, &baseAccount)
	if err != nil {
		return 0, 0, errors.Wrap(err, "fail to unmarshal base account")
	}
	return baseAccount.AccountNumber, baseAccount.Sequence, nil
}

// GetLastObservedInHeight returns the lastobservedin value for the chain past in
func (c *Client) GetLastObservedInHeight(chain common.Chain) (uint64, error) {
	lastblock, err := c.getLastBlock(chain)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to GetLastObservedInHeight")
	}
	return lastblock.LastChainHeight.Uint64(), nil
}

// GetLastSignedOutheight returns the lastsignedout value for the chain past in
func (c *Client) GetLastSignedOutHeight(chain common.Chain) (uint64, error) {
	lastblock, err := c.getLastBlock(chain)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to GetLastSignedOutheight")
	}
	return lastblock.LastSignedHeight.Uint64(), nil
}

// getLastBlock calls the /lastblock/{chain} endpoint and Unmarshal's into the QueryResHeights type
func (c *Client) getLastBlock(chain common.Chain) (stypes.QueryResHeights, error) {
	path := fmt.Sprintf("/thorchain/lastblock/%s", chain.String())
	buf, err := c.get(path)
	if err != nil {
		return stypes.QueryResHeights{}, errors.Wrap(err, "failed to get lastblock")
	}
	var lastBlock stypes.QueryResHeights
	if err := c.cdc.UnmarshalJSON(buf, &lastBlock); nil != err {
		c.errCounter.WithLabelValues("fail_unmarshal_lastblock", "").Inc()
		return stypes.QueryResHeights{}, errors.Wrap(err, "fail to unmarshal last block")
	}
	return lastBlock, nil
}

// get handle all the low level http calls using retryablehttp.Client
func (c *Client) get(path string) ([]byte, error) {
	resp, err := c.client.Get(c.getThorChainUrl(path))
	if err != nil {
		return nil, errors.Wrap(err, "fail to get from thorchain")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			c.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Status code: " + strconv.Itoa(resp.StatusCode) + " returned")
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}
	return buf, nil
}

// getThorChainUrl with the given path
func (c *Client) getThorChainUrl(path string) string {
	uri := url.URL{
		Scheme: "http",
		Host:   c.cfg.ChainHost,
		Path:   path,
	}
	return uri.String()
}

// ensureNodeWhitelistedWithTimeout run's ensureNodeWhitelisted with retry logic for a period of an hour.
func (c *Client) ensureNodeWhitelistedWithTimeout() error {
	for {
		select {
		case <-time.After(time.Hour):
			return errors.New("bifrost is not whitelisted yet")
		default:
			err := c.ensureNodeWhitelisted()
			if err == nil {
				// node had been whitelisted
				return nil
			}
			c.logger.Error().Err(err).Msg("bifrost is not whitelisted , will retry a bit later")
			time.Sleep(time.Second * 30)
		}
	}
}

// ensureNodeWhitelisted will call to thorchain to check whether the bifrost had been whitelist or not
func (c *Client) ensureNodeWhitelisted() error {
	bepAddr := c.keys.GetSignerInfo().GetAddress().String()
	if len(bepAddr) == 0 {
		return errors.New("bep address is empty")
	}
	requestUri := fmt.Sprintf("/thorchain/observer/%s", bepAddr)
	c.logger.Debug().Str("request_uri", requestUri).Msg("check node account status")
	buf, err := c.get(requestUri)
	if err != nil {
		return errors.Wrap(err, "failed to call: "+requestUri)
	}
	var nodeAccount stypes.NodeAccount
	if err := json.Unmarshal(buf, &nodeAccount); nil != err {
		c.errCounter.WithLabelValues("fail_unmarshal_nodeaccount", "").Inc()
		return errors.Wrap(err, "fail to unmarshal node account")
	}
	if nodeAccount.Status == stypes.Disabled || nodeAccount.Status == stypes.Unknown {
		return errors.Errorf("node account status %s , will not be able to forward transaction to thorchain", nodeAccount.Status)
	}
	return nil
}
