package thorclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
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
	types "gitlab.com/thorchain/thornode/bifrostv2/thorclient/types"
	ttypes "gitlab.com/thorchain/thornode/bifrostv2/types"
)

const (
	BaseEndpoint = "/thorchain"

	VaultsEndpoint      = BaseEndpoint + "/vaults/pubkeys"
	KeysignEndpoint     = BaseEndpoint + "/keysign"
	KeygenEndpoint      = BaseEndpoint + "/keygen"
	NodeAccountEndpoint = BaseEndpoint + "/nodeaccount"
	LastBlockEndpoint   = BaseEndpoint + "/lastblock"

	AuthAccountEndpoint = "/auth/accounts"
)

// Client is for all communication to a thorNode
type Client struct {
	logger              zerolog.Logger
	cdc                 *codec.Codec
	cfg                 config.ThorChainConfiguration
	keys                *keys.Keys
	pkm                 *ttypes.PubKeyManager
	errCounter          *prometheus.CounterVec
	metrics             *metrics.Metrics
	accountNumber       uint64
	seqNumber           uint64
	client              *retryablehttp.Client
	wg                  *sync.WaitGroup
	stopChan            chan struct{}
	txOutChan           chan stypes.TxOut
	keygensChan         chan stypes.Keygens
	lastSignedOutHeight int64
	backOffCtrl         backoff.ExponentialBackOff
}

// NewClient create a new instance of Client
func NewClient(cfg config.ThorChainConfiguration, m *metrics.Metrics) (*Client, error) {
	// main module logger
	logger := log.With().Str("module", "thorClient").Logger()

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
		logger.Error().Err(err).Msg("failed to get keybase")
		return nil, fmt.Errorf("failed to get keybase: %w", err)
	}

	// create retryablehttp client using our own logger format with a sublogger
	sublogger := logger.With().Str("component", "retryableHTTPClient").Logger()
	httpClientLogger := common.NewRetryableHTTPLogger(sublogger)
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = httpClientLogger

	return &Client{
		logger:      logger,
		cdc:         MakeCodec(),
		cfg:         cfg,
		keys:        k,
		errCounter:  m.GetCounterVec(metrics.ThorChainClientError),
		client:      httpClient,
		metrics:     m,
		wg:          &sync.WaitGroup{},
		stopChan:    make(chan struct{}),
		txOutChan:   make(chan stypes.TxOut),
		keygensChan: make(chan stypes.Keygens),
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

// MakeCodec used to UnmarshalJSON
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

// Start ensure that the bifrost has been whitelisted and is ready to run then
// starts to scan blocks on thorchain
func (c *Client) Start() error {
	c.logger.Info().Msg("starting thorclient")
	CosmosSDKConfig()

	// Reset/set the backOffCtrl
	c.backOffCtrl.Reset()

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

	err = c.getPubKeys()
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to retrieve pub keys")
		return errors.Wrap(err, "failed to retrieve pub keys")
	}

	// Start block scanner
	lastSignedOutHeight, err := c.GetLastSignedOutHeight()
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to retrieve last signed height on statechain")
		return errors.Wrap(err, "failed to retrieve last signed height to start block scanner")
	}
	c.logger.Info().Int64("height", lastSignedOutHeight).Msgf("last signed out height %d", lastSignedOutHeight)
	c.lastSignedOutHeight = lastSignedOutHeight
	c.wg.Add(1)
	go c.scanBlocks()

	return nil
}

func (c *Client) getPubKeys() error {
	// creates pub key manager
	pkm := ttypes.NewPubKeyManager()

	var na *stypes.NodeAccount
	for i := 0; i < 300; i++ { // wait for 5 min before timing out
		var err error
		na, err = c.GetNodeAccount(c.keys.GetSignerInfo().GetAddress().String())
		if nil != err {
			return errors.Wrap(err, "failed to get node account from thorchain")
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
		return errors.New("failed to find pubkey for this node account")
	}
	pkm.Add(na.PubKeySet.Secp256k1)
	c.pkm = pkm
	return nil
}

// Stop stops block scanner, close channels
func (c *Client) Stop() error {
	c.logger.Info().Msg("stopping thorclient")
	defer c.logger.Info().Msg("stopped thorclient successfully")
	close(c.stopChan)
	c.wg.Wait()
	return nil
}

// GetTxOutMessages return the channel with TxOut messages from thorchain
func (c *Client) GetTxOutMessages() <-chan stypes.TxOut {
	return c.txOutChan
}

// GetKeygenMessages return the channel with keygen messages from thorchain
func (c *Client) GetKeygenMessages() <-chan stypes.Keygens {
	return c.keygensChan
}

func (c *Client) scanBlocks() {
	c.logger.Info().Msg("starting scan blocks")
	defer c.logger.Info().Msg("stopped scan blocks")
	defer c.wg.Done()
	for {
		select {
		// close stop channel triggered we exit
		case <-c.stopChan:
			c.logger.Debug().Msg("stopChan closed exiting scan blocks loop")
			return
		default:
			height := c.lastSignedOutHeight
			chainHeight, err := c.GetStatechainHeight()
			if err != nil {
				c.logger.Error().Err(err).Msg("failed to get statechain height")
				time.Sleep(c.backOffCtrl.NextBackOff())
				continue
			}
			c.logger.Info().Int64("chain height", chainHeight).Msg("next block scanning")

			if height >= chainHeight {
				c.logger.Error().Int64("chain height", chainHeight).Int64("last signed height", c.lastSignedOutHeight).Msg("last signed block height is bigger or equal to chain length")
				time.Sleep(c.backOffCtrl.NextBackOff())
				continue
			}

			if height < chainHeight {
				for i := height; i < chainHeight; i++ {
					c.lastSignedOutHeight++
					c.metrics.GetCounter(metrics.TotalBlockScanned).Inc()
					err := c.processTxOutBlock(height)
					if err != nil {
						c.logger.Error().Err(err).Int64("block height", height).Msg("failed to process tx out block")
						continue
					}
					err = c.processKeygenBlock(height)
					if err != nil {
						c.logger.Error().Err(err).Int64("block height", height).Msg("failed to process keygens block")
						continue
					}
				}
			}
			c.backOffCtrl.Reset()
		}
	}
}

// processTxOutBlock retrieve txout from this block height and pass results to
// txOutChan
func (c *Client) processTxOutBlock(blockHeight int64) error {
	for _, pk := range c.pkm.GetPks() {
		if len(pk.String()) == 0 {
			continue
		}
		txOut, err := c.GetKeysign(blockHeight, pk.String())
		if err != nil {
			return err
		}
		c.txOutChan <- *txOut
	}
	return nil
}

// processKeygenBlock retrieve keygen from this block height and pass results to
// keygensChan
func (c *Client) processKeygenBlock(blockHeight int64) error {
	for _, pk := range c.pkm.GetPks() {
		if len(pk.String()) == 0 {
			continue
		}
		keygens, err := c.GetKeygens(blockHeight, pk.String())
		if err != nil {
			return err
		}
		c.keygensChan <- *keygens
	}
	return nil
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
	resp, err := c.client.Get(c.getThorChainURL(path))
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
