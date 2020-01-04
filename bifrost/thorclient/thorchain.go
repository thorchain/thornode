package thorclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

var EmptyNodeAccount stypes.NodeAccount

// ThorchainBridge will be used to send tx to thorchain
type ThorchainBridge struct {
	logger        zerolog.Logger
	cdc           *codec.Codec
	cfg           config.ThorchainConfiguration
	keys          *Keys
	errCounter    *prometheus.CounterVec
	m             *metrics.Metrics
	accountNumber uint64
	seqNumber     uint64
	client        *retryablehttp.Client
}

// NewThorchainBridge create a new instance of ThorchainBridge
func NewThorchainBridge(cfg config.ThorchainConfiguration, m *metrics.Metrics) (*ThorchainBridge, error) {
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
	return &ThorchainBridge{
		logger:     log.With().Str("module", "thorchain_bridge").Logger(),
		cdc:        MakeCodec(),
		cfg:        cfg,
		keys:       k,
		errCounter: m.GetCounterVec(metrics.ThorchainBridgeError),
		client:     retryablehttp.NewClient(),
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

func (scb *ThorchainBridge) WithRetryableHttpClient(c *retryablehttp.Client) {
	scb.client = c
}

func (scb *ThorchainBridge) Start() error {
	return nil
}

func (scb *ThorchainBridge) getAccountInfoUrl(chainHost string) string {
	return scb.GetUrl(fmt.Sprintf("/auth/accounts/%s", scb.keys.GetSignerInfo().GetAddress()))
}

func (scb *ThorchainBridge) getAccountNumberAndSequenceNumber(requestUrl string) (uint64, uint64, error) {
	if len(requestUrl) == 0 {
		return 0, 0, errors.New("request url is empty")
	}

	resp, err := scb.client.Get(requestUrl)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "fail to get response from %s", requestUrl)
	}
	if resp.StatusCode != http.StatusOK {
		return 0, 0, errors.Errorf("status code %d (%s) is unexpected", resp.StatusCode, resp.Status)
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			scb.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, errors.Wrap(err, "fail to read response body")
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

// Sign new keygen
func (scb *ThorchainBridge) GetKeygenStdTx(poolPubKey common.PubKey, inputPks common.PubKeys) (*authtypes.StdTx, error) {
	start := time.Now()
	defer func() {
		scb.m.GetHistograms(metrics.SignToThorchainDuration).Observe(time.Since(start).Seconds())
	}()

	msg := stypes.NewMsgTssPool(inputPks, poolPubKey, scb.keys.GetSignerInfo().GetAddress())

	stdTx := authtypes.NewStdTx(
		[]sdk.Msg{msg},
		authtypes.NewStdFee(100000000, nil), // fee
		nil,                                 // signatures
		"",                                  // memo
	)

	return &stdTx, nil
}

// Sign the incoming transaction
func (scb *ThorchainBridge) GetObservationsStdTx(txIns stypes.ObservedTxs) (*authtypes.StdTx, error) {
	if len(txIns) == 0 {
		scb.errCounter.WithLabelValues("nothing_to_sign", "").Inc()
		return nil, errors.New("nothing to be signed")
	}
	start := time.Now()
	defer func() {
		scb.m.GetHistograms(metrics.SignToThorchainDuration).Observe(time.Since(start).Seconds())
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
		msgs = append(msgs, stypes.NewMsgObservedTxIn(inbound, scb.keys.GetSignerInfo().GetAddress()))
	}
	if len(outbound) > 0 {
		msgs = append(msgs, stypes.NewMsgObservedTxOut(outbound, scb.keys.GetSignerInfo().GetAddress()))
	}

	stdTx := authtypes.NewStdTx(
		msgs,
		authtypes.NewStdFee(100000000, nil), // fee
		nil,                                 // signatures
		"",                                  // memo
	)

	return &stdTx, nil
}

// Send the signed transaction to thorchain
func (scb *ThorchainBridge) Send(stdTx authtypes.StdTx, mode types.TxMode) (common.TxID, error) {
	var noTxID = common.TxID("")
	if !mode.IsValid() {
		return noTxID, fmt.Errorf("transaction Mode (%s) is invalid", mode)
	}
	start := time.Now()
	defer func() {
		scb.m.GetHistograms(metrics.SendToThorchainDuration).Observe(time.Since(start).Seconds())
	}()

	accountNumber, sequenceNumber, err := scb.getAccountNumberAndSequenceNumber(scb.getAccountInfoUrl(scb.cfg.ChainHost))
	if nil != err {
		return noTxID, errors.Wrap(err, "fail to get account number and sequence number from thorchain ")
	}
	scb.logger.Info().Uint64("account_number", accountNumber).Uint64("sequence_number", sequenceNumber).Msg("account info")
	stdMsg := authtypes.StdSignMsg{
		ChainID:       scb.cfg.ChainID,
		AccountNumber: accountNumber,
		Sequence:      sequenceNumber,
		Fee:           stdTx.Fee,
		Msgs:          stdTx.GetMsgs(),
		Memo:          stdTx.GetMemo(),
	}
	sig, err := authtypes.MakeSignature(scb.keys.GetKeybase(), scb.cfg.SignerName, scb.cfg.SignerPasswd, stdMsg)
	if err != nil {
		scb.errCounter.WithLabelValues("fail_sign", "").Inc()
		return noTxID, errors.Wrap(err, "fail to sign the message")
	}

	signed := authtypes.NewStdTx(
		stdTx.GetMsgs(),
		stdTx.Fee,
		[]authtypes.StdSignature{sig},
		stdTx.GetMemo(),
	)

	scb.m.GetCounter(metrics.TxToThorchainSigned).Inc()

	var setTx types.SetTx
	setTx.Mode = mode.String()
	setTx.Tx.Msg = signed.Msgs
	setTx.Tx.Fee = signed.Fee
	setTx.Tx.Signatures = signed.Signatures
	setTx.Tx.Memo = signed.Memo
	result, err := scb.cdc.MarshalJSON(setTx)
	if nil != err {
		scb.errCounter.WithLabelValues("fail_marshal_settx", "").Inc()
		return noTxID, errors.Wrap(err, "fail to marshal settx to json")
	}

	scb.logger.Info().Str("payload", string(result)).Msg("post to thorchain")

	resp, err := scb.client.Post(scb.GetUrl("/txs"), "application/json", bytes.NewBuffer(result))
	if err != nil {
		scb.errCounter.WithLabelValues("fail_post_to_thorchain", "").Inc()
		return noTxID, errors.Wrap(err, "fail to post tx to thorchain")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			scb.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		scb.errCounter.WithLabelValues("fail_read_thorchain_resp", "").Inc()
		return noTxID, errors.Wrap(err, "fail to read response body")
	}
	var commit types.Commit
	err = json.Unmarshal(body, &commit)
	if err != nil {
		scb.errCounter.WithLabelValues("fail_unmarshal_commit", "").Inc()
		return noTxID, errors.Wrap(err, "fail to unmarshal commit")
	}
	scb.m.GetCounter(metrics.TxToThorchain).Inc()
	scb.logger.Info().Msgf("Received a TxHash of %v from the thorchain", commit.TxHash)
	return common.NewTxID(commit.TxHash)
}

// GetBinanceChainStartHeight
func (scb *ThorchainBridge) GetBinanceChainStartHeight() (int64, error) {

	resp, err := scb.client.Get(scb.GetUrl("/thorchain/lastblock"))
	if nil != err {
		return 0, errors.Wrap(err, "fail to get last blocks from thorchain")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			scb.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("fail to get last block height from thorchain")
	}
	var lastBlock stypes.QueryResHeights
	buf, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return 0, errors.Wrap(err, "fail to read response body")
	}
	if err := scb.cdc.UnmarshalJSON(buf, &lastBlock); nil != err {
		scb.errCounter.WithLabelValues("fail_unmarshal_lastblock", "").Inc()
		return 0, errors.Wrap(err, "fail to unmarshal last block")
	}

	return lastBlock.LastChainHeight, nil
}

// getThorchainUrl with the given path
func (scb *ThorchainBridge) GetUrl(path string) string {
	uri := url.URL{
		Scheme: "http",
		Host:   scb.cfg.ChainHost,
		Path:   path,
	}
	return uri.String()
}

func (scb *ThorchainBridge) EnsureNodeWhitelistedWithTimeout() error {
	for {
		select {
		case <-time.After(time.Hour):
			return errors.New("Observer is not whitelisted yet")
		default:
			err := scb.EnsureNodeWhitelisted()
			if nil == err {
				// node had been whitelisted
				return nil
			}
			scb.logger.Error().Err(err).Msg("observer is not whitelisted , will retry a bit later")
			time.Sleep(time.Second * 30)
		}
	}
}

// EnsureNodeWhitelisted will call to thorchain to check whether the observer had been whitelist or not
func (scb *ThorchainBridge) EnsureNodeWhitelisted() error {
	bepAddr := scb.keys.GetSignerInfo().GetAddress().String()
	if len(bepAddr) == 0 {
		return errors.New("bep address is empty")
	}

	requestUrl := scb.GetUrl("/thorchain/observer/" + bepAddr)
	scb.logger.Debug().Str("request_url", requestUrl).Msg("check node account status")
	resp, err := scb.client.Get(requestUrl)
	if nil != err {
		return errors.Wrap(err, "fail to get node account status")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			scb.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return errors.New("fail to get node account from thorchain")
	}
	var nodeAccount stypes.NodeAccount
	buf, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return errors.Wrap(err, "fail to read response body")
	}
	if err := scb.cdc.UnmarshalJSON(buf, &nodeAccount); nil != err {
		scb.errCounter.WithLabelValues("fail_unmarshal_nodeaccount", "").Inc()
		return errors.Wrap(err, "fail to unmarshal node account")
	}

	if nodeAccount.Status == stypes.Disabled || nodeAccount.Status == stypes.Unknown {
		return errors.Errorf("node account status %s , will not be able to forward transaction to thorchain", nodeAccount.Status)
	}
	return nil
}

// GetNodeAccount from thorchain
func (scb *ThorchainBridge) GetNodeAccount(thorAddr string) (stypes.NodeAccount, error) {
	requestUrl := scb.GetUrl("/thorchain/nodeaccount/" + thorAddr)

	scb.logger.Debug().Str("request_url", requestUrl).Msg("get node account")
	resp, err := scb.client.Get(requestUrl)
	if nil != err {
		return EmptyNodeAccount, errors.Wrap(err, "fail to get node account")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			scb.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return EmptyNodeAccount, fmt.Errorf("fail to get node account from thorchain,statusCode:%d", resp.StatusCode)
	}
	var na stypes.NodeAccount

	buf, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return EmptyNodeAccount, fmt.Errorf("fail to read response body,err:%w", err)
	}
	cdc := MakeCodec()
	if err := cdc.UnmarshalJSON(buf, &na); nil != err {
		return EmptyNodeAccount, fmt.Errorf("fail to unmarshal node account response,err:%w", err)
	}
	return na, nil
}
