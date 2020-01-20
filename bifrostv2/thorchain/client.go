package thorchain

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
	"gitlab.com/thorchain/thornode/common"
	stypes "gitlab.com/thorchain/thornode/x/thorchain/types"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/metrics"
	types "gitlab.com/thorchain/thornode/bifrostv2/thorchain/types"
	ttypes "gitlab.com/thorchain/thornode/bifrostv2/types"
)

// Endpoint urls
const (
	AuthAccountEndpoint = "/auth/accounts"
	KeygenEndpoint      = "/thorchain/keygen"
	KeysignEndpoint     = "/thorchain/keysign"
	LastBlockEndpoint   = "/thorchain/lastblock"
	NodeAccountEndpoint = "/thorchain/nodeaccount"
	VaultsEndpoint      = "/thorchain/vaults/pubkeys"
)

// Client is for all communication to a thorNode
type Client struct {
	logger        zerolog.Logger
	cdc           *codec.Codec
	cfg           config.ClientConfiguration
	keys          *keys.Keys
	errCounter    *prometheus.CounterVec
	metrics       *metrics.Metrics
	accountNumber uint64
	seqNumber     uint64
	httpClient    *retryablehttp.Client
}

// NewClient create a new instance of Client
func NewClient(cfg config.ClientConfiguration, m *metrics.Metrics) (*Client, error) {
	// main module logger
	logger := log.With().Str("module", "thorchain_client").Logger()

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
	if err != nil {
		logger.Error().Err(err).Msg("failed to get keybase")
		return nil, errors.Wrap(err, "failed to get keybase")
	}

	// create retryablehttp client using our own logger format with a sublogger
	sublogger := logger.With().Str("component", "retryable_http_client").Logger()
	httpClientLogger := common.NewRetryableHTTPLogger(sublogger)
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = httpClientLogger

	return &Client{
		logger:     logger,
		cdc:        MakeCodec(),
		cfg:        cfg,
		keys:       k,
		errCounter: m.GetCounterVec(metrics.ThorchainClientError),
		httpClient: httpClient,
		metrics:    m,
	}, nil
}

// MakeCodec used to UnmarshalJSON
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	sdk.RegisterCodec(cdc)
	stypes.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

// Start ensure that the bifrost has been whitelisted and is ready to run
func (c *Client) Start() error {
	c.logger.Info().Msg("starting thorchain client")

	if err := c.ensureNodeWhitelistedWithTimeout(); err != nil {
		c.logger.Error().Err(err).Msg("node account is not whitelisted, can't start")
		return errors.Wrap(err, "node account is not whitelisted, can't start")
	}

	accountNumber, sequenceNumber, err := c.getAccountNumberAndSequenceNumber()
	if nil != err {
		c.logger.Error().Err(err).Msg("failed to get account number and sequence number from thorchain")
		return errors.Wrap(err, "failed to get account number and sequence number from thorchain")
	}

	c.logger.Info().Uint64("account number", accountNumber).Uint64("sequence no", sequenceNumber).Msg("account information")
	c.accountNumber = accountNumber
	c.seqNumber = sequenceNumber

	return nil
}

// Stop stops client, nothing to do here
func (c *Client) Stop() error {
	c.logger.Info().Msg("stopped thorchain client")
	return nil
}

// GetPubKeys retrieves pub keys for this node (yggdrasil + asgard)
func (c *Client) GetPubKeys() (*ttypes.PubKeyManager, error) {
	// creates pub key manager
	pkm := ttypes.NewPubKeyManager()

	var na *stypes.NodeAccount
	for i := 0; i < 300; i++ { // wait for 5 min before timing out
		var err error
		na, err = c.GetNodeAccount(c.keys.GetSignerInfo().GetAddress().String())
		if nil != err {
			return &ttypes.PubKeyManager{}, errors.Wrap(err, "failed to get node account from thorchain")
		}

		if !na.PubKeySet.Secp256k1.IsEmpty() {
			break
		}
		time.Sleep(5 * time.Second)

		c.logger.Info().Msg("Waiting for node account to be registered...")
	}
	for _, item := range na.SignerMembership {
		pkm.Add(item)
	}

	if na.PubKeySet.Secp256k1.IsEmpty() {
		return &ttypes.PubKeyManager{}, errors.New("failed to find pubkey for this node account")
	}
	pkm.Add(na.PubKeySet.Secp256k1)
	return pkm, nil
}

// getAccountNumberAndSequenceNumber returns account and Sequence number required to post into thorchain
func (c *Client) getAccountNumberAndSequenceNumber() (uint64, uint64, error) {
	url := fmt.Sprintf("%s/%s", AuthAccountEndpoint, c.keys.GetSignerInfo().GetAddress())

	body, err := c.get(url)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to get auth accounts")
	}
	var accountResp types.AccountResp
	if err := json.Unmarshal(body, &accountResp); nil != err {
		return 0, 0, errors.Wrap(err, "failed to unmarshal account resp")
	}
	var baseAccount authtypes.BaseAccount
	err = authtypes.ModuleCdc.UnmarshalJSON(accountResp.Result, &baseAccount)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to unmarshal base account")
	}
	return baseAccount.AccountNumber, baseAccount.Sequence, nil
}

// GetLastObservedInHeight returns the lastobservedin value for the chain past in
func (c *Client) GetLastObservedInHeight(chain common.Chain) (int64, error) {
	lastblock, err := c.getLastBlock(chain)
	if err != nil {
		return 0, errors.Wrap(err, "failed to GetLastObservedInHeight")
	}
	return lastblock.LastChainHeight, nil
}

// GetLastSignedOutHeight returns the lastsignedout value for thorchain
func (c *Client) GetLastSignedOutHeight() (int64, error) {
	lastblock, err := c.getLastBlock("")
	if err != nil {
		return 0, errors.Wrap(err, "failed to GetLastSignedOutheight")
	}
	return lastblock.LastSignedHeight, nil
}

// GetStatechainHeight returns the current height for thorchain blocks
func (c *Client) GetStatechainHeight() (int64, error) {
	lastblock, err := c.getLastBlock("")
	if err != nil {
		return 0, errors.Wrap(err, "failed to GetStatechainHeight")
	}
	return lastblock.Statechain, nil
}

// getLastBlock calls the /lastblock/{chain} endpoint and Unmarshal's into the QueryResHeights type
func (c *Client) getLastBlock(chain common.Chain) (stypes.QueryResHeights, error) {
	url := fmt.Sprintf("%s/%s", LastBlockEndpoint, chain.String())
	buf, err := c.get(url)
	if err != nil {
		return stypes.QueryResHeights{}, errors.Wrap(err, "failed to get lastblock")
	}
	var lastBlock stypes.QueryResHeights
	if err := c.cdc.UnmarshalJSON(buf, &lastBlock); nil != err {
		c.errCounter.WithLabelValues("fail_unmarshal_lastblock", "").Inc()
		return stypes.QueryResHeights{}, errors.Wrap(err, "failed to unmarshal last block")
	}
	return lastBlock, nil
}

// get handle all the low level http calls using retryablehttp.Client
func (c *Client) get(path string) ([]byte, error) {
	resp, err := c.httpClient.Get(c.getThorChainURL(path))
	if err != nil {
		return nil, errors.Wrap(err, "failed to GET from thorchain")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			c.logger.Error().Err(err).Msg("failed to close response body")
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

// getThorChainURL with the given path
func (c *Client) getThorChainURL(path string) string {
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
	na, err := c.GetNodeAccount(bepAddr)
	if err != nil {
		return errors.Wrap(err, "failed to get node account")
	}
	if na.Status == stypes.Disabled || na.Status == stypes.Unknown {
		return errors.Errorf("node account status %s , will not be able to forward transaction to thorchain", na.Status)
	}
	return nil
}
