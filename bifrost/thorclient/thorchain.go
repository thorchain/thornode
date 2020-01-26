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

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/common"
	stypes "gitlab.com/thorchain/thornode/x/thorchain/types"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/thorclient/types"
)

// Endpoint urls
const (
	AuthAccountEndpoint  = "/auth/accounts"
	BroadcastTxsEndpoint = "/txs"
	KeygenEndpoint       = "/thorchain/keygen"
	KeysignEndpoint      = "/thorchain/keysign"
	LastBlockEndpoint    = "/thorchain/lastblock"
	NodeAccountEndpoint  = "/thorchain/nodeaccount"
	ValidatorsEndpoint   = "/thorchain/validators"
	VaultsEndpoint       = "/thorchain/vaults/pubkeys"
)

// ThorchainBridge will be used to send tx to thorchain
type ThorchainBridge struct {
	logger        zerolog.Logger
	cdc           *codec.Codec
	cfg           config.ClientConfiguration
	keys          *Keys
	errCounter    *prometheus.CounterVec
	m             *metrics.Metrics
	blockHeight   int64
	accountNumber uint64
	seqNumber     uint64
	httpClient    *retryablehttp.Client
	broadcastLock *sync.RWMutex
}

// NewThorchainBridge create a new instance of ThorchainBridge
func NewThorchainBridge(cfg config.ClientConfiguration, m *metrics.Metrics) (*ThorchainBridge, error) {
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
	k, err := NewKeys(cfg.ChainHomeFolder, cfg.SignerName, cfg.SignerPasswd)
	if nil != err {
		return nil, fmt.Errorf("fail to get keybase,err:%w", err)
	}

	// create retryablehttp client using our own logger format with a sublogger
	sublogger := logger.With().Str("component", "retryable_http_client").Logger()
	httpClientLogger := common.NewRetryableHTTPLogger(sublogger)
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = httpClientLogger

	return &ThorchainBridge{
		logger:        logger,
		cdc:           MakeCodec(),
		cfg:           cfg,
		keys:          k,
		errCounter:    m.GetCounterVec(metrics.ThorchainClientError),
		httpClient:    httpClient,
		m:             m,
		broadcastLock: &sync.RWMutex{},
	}, nil
}

// MakeCodec creates codec
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	sdk.RegisterCodec(cdc)
	stypes.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

// get handle all the low level http GET calls using retryablehttp.ThorchainBridge
func (b *ThorchainBridge) get(path string) ([]byte, error) {
	resp, err := b.httpClient.Get(b.getThorChainURL(path))
	if err != nil {
		b.errCounter.WithLabelValues("fail_get_from_thorchain", "").Inc()
		return nil, errors.Wrap(err, "failed to GET from thorchain")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			b.logger.Error().Err(err).Msg("failed to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Status code: " + strconv.Itoa(resp.StatusCode) + " returned")
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		b.errCounter.WithLabelValues("fail_read_thorchain_resp", "").Inc()
		return nil, errors.Wrap(err, "failed to read response body")
	}
	return buf, nil
}

// post handle all the low level http POST calls using retryablehttp.ThorchainBridge
func (b *ThorchainBridge) post(path string, bodyType string, body interface{}) ([]byte, error) {
	resp, err := b.httpClient.Post(b.getThorChainURL(path), bodyType, body)
	if err != nil {
		b.errCounter.WithLabelValues("fail_post_to_thorchain", "").Inc()
		return nil, errors.Wrap(err, "failed to POST to thorchain")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			b.logger.Error().Err(err).Msg("failed to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Status code: " + strconv.Itoa(resp.StatusCode) + " returned")
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		b.errCounter.WithLabelValues("fail_read_thorchain_resp", "").Inc()
		return nil, errors.Wrap(err, "failed to read response body")
	}
	return buf, nil
}

// getThorChainURL with the given path
func (b *ThorchainBridge) getThorChainURL(path string) string {
	uri := url.URL{
		Scheme: "http",
		Host:   b.cfg.ChainHost,
		Path:   path,
	}
	return uri.String()
}

// getAccountNumberAndSequenceNumber returns account and Sequence number required to post into thorchain
func (b *ThorchainBridge) getAccountNumberAndSequenceNumber() (uint64, uint64, error) {
	url := fmt.Sprintf("%s/%s", AuthAccountEndpoint, b.keys.GetSignerInfo().GetAddress())

	body, err := b.get(url)
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

// GetKeygenStdTx get keygen tx from params
func (b *ThorchainBridge) GetKeygenStdTx(poolPubKey common.PubKey, inputPks common.PubKeys) (*authtypes.StdTx, error) {
	start := time.Now()
	defer func() {
		b.m.GetHistograms(metrics.SignToThorchainDuration).Observe(time.Since(start).Seconds())
	}()

	msg := stypes.NewMsgTssPool(inputPks, poolPubKey, b.keys.GetSignerInfo().GetAddress())

	stdTx := authtypes.NewStdTx(
		[]sdk.Msg{msg},
		authtypes.NewStdFee(100000000, nil), // fee
		nil,                                 // signatures
		"",                                  // memo
	)

	return &stdTx, nil
}

// GetObservationsStdTx get observations tx from txIns
func (b *ThorchainBridge) GetObservationsStdTx(txIns stypes.ObservedTxs) (*authtypes.StdTx, error) {
	if len(txIns) == 0 {
		b.errCounter.WithLabelValues("nothing_to_sign", "").Inc()
		return nil, errors.New("nothing to be signed")
	}
	start := time.Now()
	defer func() {
		b.m.GetHistograms(metrics.SignToThorchainDuration).Observe(time.Since(start).Seconds())
	}()

	var inbound stypes.ObservedTxs
	var outbound stypes.ObservedTxs

	// spilt our txs into inbound vs outbound txs
	for _, tx := range txIns {
		chain := common.BNBChain
		if len(tx.Tx.Coins) > 0 {
			chain = tx.Tx.Coins[0].Asset.Chain
		}

		obAddr, err := tx.ObservedPubKey.GetAddress(chain)
		if err != nil {
			return nil, err
		}
		if tx.Tx.ToAddress.Equals(obAddr) {
			inbound = append(inbound, tx)
		} else if tx.Tx.FromAddress.Equals(obAddr) {
			outbound = append(outbound, tx)
		} else {
			return nil, errors.New("Could not determine if this tx as inbound or outbound")
		}
	}

	var msgs []sdk.Msg
	if len(inbound) > 0 {
		msgs = append(msgs, stypes.NewMsgObservedTxIn(inbound, b.keys.GetSignerInfo().GetAddress()))
	}
	if len(outbound) > 0 {
		msgs = append(msgs, stypes.NewMsgObservedTxOut(outbound, b.keys.GetSignerInfo().GetAddress()))
	}

	stdTx := authtypes.NewStdTx(
		msgs,
		authtypes.NewStdFee(100000000, nil), // fee
		nil,                                 // signatures
		"",                                  // memo
	)

	return &stdTx, nil
}

// EnsureNodeWhitelistedWithTimeout check node is whitelisted with timeout retry
func (b *ThorchainBridge) EnsureNodeWhitelistedWithTimeout() error {
	for {
		select {
		case <-time.After(time.Hour):
			return errors.New("Observer is not whitelisted yet")
		default:
			err := b.EnsureNodeWhitelisted()
			if nil == err {
				// node had been whitelisted
				return nil
			}
			b.logger.Error().Err(err).Msg("observer is not whitelisted , will retry a bit later")
			time.Sleep(time.Second * 30)
		}
	}
}

// EnsureNodeWhitelisted will call to thorchain to check whether the observer had been whitelist or not
func (b *ThorchainBridge) EnsureNodeWhitelisted() error {
	bepAddr := b.keys.GetSignerInfo().GetAddress().String()
	if len(bepAddr) == 0 {
		return errors.New("bep address is empty")
	}
	na, err := b.GetNodeAccount(bepAddr)
	if err != nil {
		return errors.Wrap(err, "failed to get node account")
	}
	if na.Status == stypes.Disabled || na.Status == stypes.Unknown {
		return errors.Errorf("node account status %s , will not be able to forward transaction to thorchain", na.Status)
	}
	return nil
}
