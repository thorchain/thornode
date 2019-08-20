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
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	ckeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/thorchain/bepswap/common"
	stypes "gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"

	"gitlab.com/thorchain/bepswap/observe/config"
	"gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

const (
	// folder name for statechain sscli
	StatechainCliFolderName = `.sscli`
)

// StateChainBridge will be used to send tx to statechain
type StateChainBridge struct {
	logger     zerolog.Logger
	cdc        *codec.Codec
	cfg        config.StateChainConfiguration
	signerInfo ckeys.Info
	kb         ckeys.Keybase
}

// NewStateChainBridge create a new instance of StateChainBridge
func NewStateChainBridge(cfg config.StateChainConfiguration) (*StateChainBridge, error) {
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
	kb, err := getKeybase(cfg.ChainHomeFolder, cfg.SignerName)
	if nil != err {
		return nil, errors.Wrap(err, "fail to get keybase")
	}
	signerInfo, err := kb.Get(cfg.SignerName)
	if nil != err {
		return nil, errors.Wrap(err, "fail to get signer info")
	}

	return &StateChainBridge{
		logger:     log.With().Str("module", "statechain_bridge").Logger(),
		cdc:        makeCodec(),
		cfg:        cfg,
		signerInfo: signerInfo,
		kb:         kb,
	}, nil
}

func makeCodec() *codec.Codec {
	var cdc = codec.New()
	sdk.RegisterCodec(cdc)
	// TODO make we should share this with statechain in common
	cdc.RegisterConcrete(stypes.MsgSetTxIn{}, "swapservice/MsgSetTxIn", nil)
	codec.RegisterCrypto(cdc)
	return cdc
}

func getKeybase(stateChainHome, signerName string) (ckeys.Keybase, error) {
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
func (scb *StateChainBridge) getAccountInfoUrl(chainHost string) string {
	uri := url.URL{
		Scheme: "http",
		Host:   chainHost,
		Path:   fmt.Sprintf("/auth/accounts/%s", scb.signerInfo.GetAddress()),
	}
	return uri.String()
}

func (scb *StateChainBridge) getAccountNumberAndSequenceNumber(requestUrl string) (int64, int64, error) {
	if len(requestUrl) == 0 {
		return 0, 0, errors.New("request url is empty")
	}
	resp, err := retryablehttp.Get(requestUrl)
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

	var baseAccount types.BaseAccount
	err = json.Unmarshal(body, &baseAccount)
	if err != nil {
		return 0, 0, errors.Wrap(err, "fail to unmarshal base account")
	}
	base := baseAccount.Value
	acctNumber, err := strconv.ParseInt(base.AccountNumber, 10, 64)
	if nil != err {
		return 0, 0, errors.Wrapf(err, "fail to parse AccountNumber(%s) to int", base.AccountNumber)
	}
	seq, err := strconv.ParseInt(base.Sequence, 10, 64)
	if nil != err {
		return 0, 0, errors.Wrapf(err, "fail to parse sequence(%s) to int", base.Sequence)
	}
	return acctNumber, seq, nil

}

// Sign the incoming transaction
func (scb *StateChainBridge) Sign(txIns []stypes.TxIn) (*authtypes.StdTx, error) {
	if len(txIns) == 0 {
		return nil, errors.New("nothing to be signed")
	}
	stdTx := authtypes.NewStdTx(
		[]sdk.Msg{
			stypes.NewMsgSetTxIn(txIns, scb.signerInfo.GetAddress()),
		}, // messages
		authtypes.NewStdFee(200000, nil), // fee
		nil,                              // signatures
		"",                               // memo
	)

	accNumber, seqNumber, err := scb.getAccountNumberAndSequenceNumber(scb.getAccountInfoUrl(scb.cfg.ChainHost))
	if nil != err {
		return nil, errors.Wrap(err, "fail to get account number and sequence number from statechain")
	}
	stdMsg := authtypes.StdSignMsg{
		ChainID:       scb.cfg.ChainID,
		AccountNumber: uint64(accNumber),
		Sequence:      uint64(seqNumber),
		Fee:           stdTx.Fee,
		Msgs:          stdTx.GetMsgs(),
		Memo:          stdTx.GetMemo(),
	}
	sig, err := authtypes.MakeSignature(scb.kb, scb.cfg.SignerName, scb.cfg.SignerPasswd, stdMsg)
	if err != nil {
		return nil, errors.Wrap(err, "fail to sign the message")
	}

	signedStdTx := authtypes.NewStdTx(
		stdTx.GetMsgs(),
		stdTx.Fee,
		[]authtypes.StdSignature{sig},
		stdTx.GetMemo(),
	)
	return &signedStdTx, nil
}

// Send the signed transaction to statechain
func (scb *StateChainBridge) Send(signed authtypes.StdTx, mode types.TxMode) (common.TxID, error) {
	var noTxID = common.TxID("")
	if !mode.IsValid() {
		return noTxID, fmt.Errorf("transaction Mode (%s) is invalid", mode)
	}

	var setTx types.SetTx
	setTx.Mode = mode.String()
	setTx.Tx.Msg = signed.Msgs
	setTx.Tx.Fee = signed.Fee
	setTx.Tx.Signatures = signed.Signatures
	setTx.Tx.Memo = signed.Memo
	result, err := scb.cdc.MarshalJSON(setTx)
	if nil != err {
		return noTxID, errors.Wrap(err, "fail to marsh settx to json")
	}
	uri := url.URL{
		Scheme: "http",
		Host:   scb.cfg.ChainHost,
		Path:   "/txs",
	}
	scb.logger.Debug().Str("payload", string(result)).Msg("post to statechain")

	resp, err := retryablehttp.Post(uri.String(), "application/json", bytes.NewBuffer(result))
	if err != nil {
		return noTxID, errors.Wrap(err, "fail to post tx to statechain")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			scb.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return noTxID, errors.Wrap(err, "fail to read response body")
	}
	var commit types.Commit
	err = json.Unmarshal(body, &commit)
	if err != nil {
		return noTxID, errors.Wrap(err, "fail to unmarshal commit")
	}

	scb.logger.Info().Msgf("Received a TxHash of %v from the statechain", commit.TxHash)
	return common.NewTxID(commit.TxHash)
}
