package statechain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/user"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	ckeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/bepswap/thornode/common"
	stypes "gitlab.com/thorchain/bepswap/thornode/x/swapservice/types"

	"gitlab.com/thorchain/bepswap/thornode/config"
	"gitlab.com/thorchain/bepswap/thornode/x/metrics"
	"gitlab.com/thorchain/bepswap/thornode/x/statechain/types"
)

const (
	// folder name for statechain thorcli
	StatechainCliFolderName = `.thorcli`
)

// StateChainBridge will be used to send tx to statechain
type StateChainBridge struct {
	logger        zerolog.Logger
	cdc           *codec.Codec
	cfg           config.StateChainConfiguration
	signerInfo    ckeys.Info
	kb            ckeys.Keybase
	errCounter    *prometheus.CounterVec
	m             *metrics.Metrics
	accountNumber uint64
	seqNumber     uint64
	client        *retryablehttp.Client
}

// NewStateChainBridge create a new instance of StateChainBridge
func NewStateChainBridge(cfg config.StateChainConfiguration, m *metrics.Metrics) (*StateChainBridge, error) {
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
	kb, err := GetKeybase(cfg.ChainHomeFolder, cfg.SignerName)
	if nil != err {
		return nil, errors.Wrap(err, "fail to get keybase")
	}
	signerInfo, err := kb.Get(cfg.SignerName)
	if nil != err {
		return nil, errors.Wrap(err, "fail to get signer info")
	}

	return &StateChainBridge{
		logger:     log.With().Str("module", "statechain_bridge").Logger(),
		cdc:        MakeCodec(),
		cfg:        cfg,
		signerInfo: signerInfo,
		kb:         kb,
		errCounter: m.GetCounterVec(metrics.StateChainBridgeError),
		client:     retryablehttp.NewClient(),
		m:          m,
	}, nil
}

func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	sdk.RegisterCodec(cdc)
	// TODO make we should share this with statechain in common
	cdc.RegisterConcrete(stypes.MsgSetTxIn{}, "swapservice/MsgSetTxIn", nil)
	codec.RegisterCrypto(cdc)
	return cdc
}

func GetKeybase(stateChainHome, signerName string) (ckeys.Keybase, error) {
	cliDir := stateChainHome
	if len(stateChainHome) == 0 {
		usr, err := user.Current()
		if nil != err {
			return nil, errors.Wrap(err, "fail to get current user")
		}
		cliDir = filepath.Join(usr.HomeDir, StatechainCliFolderName)
	}
	return keys.NewKeyBaseFromDir(cliDir)
}

func (scb *StateChainBridge) WithRetryableHttpClient(c *retryablehttp.Client) {
	scb.client = c
}

func (scb *StateChainBridge) Start() error {
	accountNumber, sequenceNumber, err := scb.getAccountNumberAndSequenceNumber(scb.getAccountInfoUrl(scb.cfg.ChainHost))
	if nil != err {
		return errors.Wrap(err, "fail to get account number and sequence number from statechain ")
	}

	scb.logger.Info().Uint64("account number", accountNumber).Uint64("sequence no", sequenceNumber).Msg("account information")
	scb.accountNumber = accountNumber
	scb.seqNumber = sequenceNumber
	return nil
}

func (scb *StateChainBridge) getAccountInfoUrl(chainHost string) string {
	return scb.getStateChainUrl(fmt.Sprintf("/auth/accounts/%s", scb.signerInfo.GetAddress()))
}

func (scb *StateChainBridge) getAccountNumberAndSequenceNumber(requestUrl string) (uint64, uint64, error) {
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

// Sign the incoming transaction
func (scb *StateChainBridge) Sign(txIns []stypes.TxInVoter) (*authtypes.StdTx, error) {
	if len(txIns) == 0 {
		scb.errCounter.WithLabelValues("nothing_to_sign", "").Inc()
		return nil, errors.New("nothing to be signed")
	}
	start := time.Now()
	defer func() {
		scb.m.GetHistograms(metrics.SignToStateChainDuration).Observe(time.Since(start).Seconds())
	}()
	stdTx := authtypes.NewStdTx(
		[]sdk.Msg{
			stypes.NewMsgSetTxIn(txIns, scb.signerInfo.GetAddress()),
		}, // messages
		authtypes.NewStdFee(100000000, nil), // fee
		nil,                                 // signatures
		"",                                  // memo
	)

	scb.logger.Info().Str("chainid", scb.cfg.ChainID).Uint64("accountnumber", scb.accountNumber).Uint64("sequenceNo", scb.seqNumber).Msg("info")
	stdMsg := authtypes.StdSignMsg{
		ChainID:       scb.cfg.ChainID,
		AccountNumber: scb.accountNumber,
		Sequence:      scb.seqNumber,
		Fee:           stdTx.Fee,
		Msgs:          stdTx.GetMsgs(),
		Memo:          stdTx.GetMemo(),
	}
	sig, err := authtypes.MakeSignature(scb.kb, scb.cfg.SignerName, scb.cfg.SignerPasswd, stdMsg)
	if err != nil {
		scb.errCounter.WithLabelValues("fail_sign", "").Inc()
		return nil, errors.Wrap(err, "fail to sign the message")
	}

	signedStdTx := authtypes.NewStdTx(
		stdTx.GetMsgs(),
		stdTx.Fee,
		[]authtypes.StdSignature{sig},
		stdTx.GetMemo(),
	)
	nextSeq := atomic.AddUint64(&scb.seqNumber, 1)
	scb.logger.Info().Uint64("sequence no", nextSeq).Msg("next sequence no")
	scb.m.GetCounter(metrics.TxToStateChainSigned).Inc()
	return &signedStdTx, nil
}

// Send the signed transaction to statechain
func (scb *StateChainBridge) Send(signed authtypes.StdTx, mode types.TxMode) (common.TxID, error) {
	var noTxID = common.TxID("")
	if !mode.IsValid() {
		return noTxID, fmt.Errorf("transaction Mode (%s) is invalid", mode)
	}
	start := time.Now()
	defer func() {
		scb.m.GetHistograms(metrics.SendToStatechainDuration).Observe(time.Since(start).Seconds())
	}()
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
	scb.logger.Info().Str("payload", string(result)).Msg("post to statechain")

	resp, err := scb.client.Post(scb.getStateChainUrl("/txs"), "application/json", bytes.NewBuffer(result))
	if err != nil {
		scb.errCounter.WithLabelValues("fail_post_to_statechain", "").Inc()
		return noTxID, errors.Wrap(err, "fail to post tx to statechain")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			scb.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		scb.errCounter.WithLabelValues("fail_read_statechain_resp", "").Inc()
		return noTxID, errors.Wrap(err, "fail to read response body")
	}
	var commit types.Commit
	err = json.Unmarshal(body, &commit)
	if err != nil {
		scb.errCounter.WithLabelValues("fail_unmarshal_commit", "").Inc()
		return noTxID, errors.Wrap(err, "fail to unmarshal commit")
	}
	scb.m.GetCounter(metrics.TxToStateChain).Inc()
	scb.logger.Info().Msgf("Received a TxHash of %v from the statechain", commit.TxHash)
	return common.NewTxID(commit.TxHash)
}

// GetBinanceChainStartHeight
func (scb *StateChainBridge) GetBinanceChainStartHeight() (uint64, error) {

	resp, err := scb.client.Get(scb.getStateChainUrl("/swapservice/lastblock"))
	if nil != err {
		return 0, errors.Wrap(err, "fail to get last blocks from statechain")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			scb.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("fail to get last block height from statechain")
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

	return lastBlock.LastChainHeight.Uint64(), nil
}

// getStateChainUrl with the given path
func (scb *StateChainBridge) getStateChainUrl(path string) string {
	uri := url.URL{
		Scheme: "http",
		Host:   scb.cfg.ChainHost,
		Path:   path,
	}
	return uri.String()
}

func (scb *StateChainBridge) EnsureNodeWhitelistedWithTimeout() error {
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

// EnsureNodeWhitelisted will call to statechain to check whether the observer had been whitelist or not
func (scb *StateChainBridge) EnsureNodeWhitelisted() error {
	bepAddr := scb.signerInfo.GetAddress().String()
	if len(bepAddr) == 0 {
		return errors.New("bep address is empty")
	}

	requestUrl := scb.getStateChainUrl("/swapservice/observer/" + bepAddr)
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
		return errors.New("fail to get node account from statechain")
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
		return errors.Errorf("node account status %s , will not be able to forward transaction to statechain", nodeAccount.Status)
	}
	return nil
}
