package tss

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"time"

	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/btcd/btcec"
	"github.com/tendermint/tendermint/crypto"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/common"
)

// KeySign is a proxy between signer and TSS
type KeySign struct {
	cfg    config.TSSConfiguration
	logger zerolog.Logger
	client *http.Client
}

// NewKeySign create a new instance of KeySign
func NewKeySign(cfg config.TSSConfiguration) (*KeySign, error) {

	if len(cfg.Host) == 0 {
		return nil, errors.New("TSS host is empty")
	}
	if cfg.Port == 0 {
		return nil, errors.New("TSS port not specified")
	}

	return &KeySign{
		cfg:    cfg,
		logger: log.With().Str("module", "tss_signer").Logger(),
		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}, nil
}

// GetPrivKey THORNode don't actually have any private key , but just return something
func (s *KeySign) GetPrivKey() crypto.PrivKey {
	return nil
}

func (s *KeySign) GetAddr() ctypes.AccAddress {
	return nil
}

// ExportAsMnemonic THORNode don't need this function for TSS, just keep it to fulfill KeyManager interface
func (s *KeySign) ExportAsMnemonic() (string, error) {
	return "", nil
}

// ExportAsPrivateKey THORNode don't need this function for TSS, just keep it to fulfill KeyManager interface
func (s *KeySign) ExportAsPrivateKey() (string, error) {
	return "", nil
}

// ExportAsKeyStore THORNode don't need this function for TSS, just keep it to fulfill KeyManager interface
func (s *KeySign) ExportAsKeyStore(password string) (*keys.EncryptedKeyJSON, error) {
	return nil, nil
}
func (s *KeySign) makeSignature(msg tx.StdSignMsg, poolPubKey string) (sig tx.StdSignature, err error) {
	var stdSignature tx.StdSignature
	pk, err := sdk.GetAccPubKeyBech32(poolPubKey)
	if nil != err {
		return stdSignature, fmt.Errorf("fail to get pub key: %w", err)
	}
	signPack, err := s.remoteSign(msg.Bytes(), poolPubKey)
	if err != nil {
		return stdSignature, errors.Wrap(err, "fail to TSS sign")
	}
	if pk.VerifyBytes(msg.Bytes(), signPack) {
		s.logger.Info().Msg("we can successfully verify the bytes")
	} else {
		s.logger.Error().Msg("Oops! we cannot verify the bytes")
	}

	return tx.StdSignature{
		AccountNumber: msg.AccountNumber,
		Sequence:      msg.Sequence,
		PubKey:        pk,
		Signature:     signPack,
	}, nil
}
func (s *KeySign) Sign(msg tx.StdSignMsg) ([]byte, error) {
	return nil, nil
}

func (s *KeySign) SignWithPool(msg tx.StdSignMsg, poolPubKey common.PubKey) ([]byte, error) {
	sig, err := s.makeSignature(msg, poolPubKey.String())
	if err != nil {
		return nil, err
	}
	newTx := tx.NewStdTx(msg.Msgs, []tx.StdSignature{sig}, msg.Memo, msg.Source, msg.Data)
	bz, err := tx.Cdc.MarshalBinaryLengthPrefixed(&newTx)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func (s *KeySign) remoteSign(msg []byte, poolPubKey string) ([]byte, error) {
	if len(msg) == 0 {
		return nil, nil
	}
	encodedMsg := base64.StdEncoding.EncodeToString(msg)
	rResult, sResult, err := s.toLocalTSSSigner(poolPubKey, encodedMsg)
	if nil != err {
		return nil, errors.Wrap(err, "fail to tss sign")
	}

	s.logger.Debug().Str("R", rResult).Str("S", sResult).Msg("tss result")
	data, err := getSignature(rResult, sResult)
	if nil != err {
		return nil, errors.Wrap(err, "fail to decode tss signature")
	}

	return data, nil
}
func getSignature(r, s string) ([]byte, error) {
	rBytes, err := base64.StdEncoding.DecodeString(r)
	if nil != err {
		return nil, err
	}
	sBytes, err := base64.StdEncoding.DecodeString(s)
	if nil != err {
		return nil, err
	}

	R := new(big.Int).SetBytes(rBytes)
	S := new(big.Int).SetBytes(sBytes)
	N := btcec.S256().N
	halfOrder := new(big.Int).Rsh(N, 1)
	// see: https://github.com/ethereum/go-ethereum/blob/f9401ae011ddf7f8d2d95020b7446c17f8d98dc1/crypto/signature_nocgo.go#L90-L93
	if S.Cmp(halfOrder) == 1 {
		S.Sub(N, S)
	}

	// Serialize signature to R || S.
	// R, S are padded to 32 bytes respectively.
	rBytes = R.Bytes()
	sBytes = S.Bytes()

	sigBytes := make([]byte, 64)
	// 0 pad the byte arrays from the left if they aren't big enough.
	copy(sigBytes[32-len(rBytes):32], rBytes)
	copy(sigBytes[64-len(sBytes):64], sBytes)
	return sigBytes, nil
}

func (s *KeySign) getTSSLocalUrl() string {
	u := url.URL{
		Scheme: s.cfg.Scheme,
		Host:   fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port),
		Path:   "keysign",
	}
	return u.String()
}

// toLocalTSSSigner will send the request to local signer
func (s *KeySign) toLocalTSSSigner(poolPubKey, sendmsg string) (string, string, error) {
	tssMsg := struct {
		PoolPubKey string `json:"pool_pub_key"`
		Message    string `json:"message"`
	}{
		PoolPubKey: poolPubKey,
		Message:    sendmsg,
	}
	buf, err := json.Marshal(tssMsg)
	if nil != err {
		return "", "", errors.Wrap(err, "fail to create tss request msg")
	}
	s.logger.Debug().Str("payload", string(buf)).Msg("msg to tss Local node")
	localTssURL := s.getTSSLocalUrl()
	resp, err := s.client.Post(localTssURL, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return "", "", errors.Wrapf(err, "fail to send request to local TSS node,url: %s", localTssURL)
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			s.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("response status: %s from tss sign ", resp.Status)
	}

	// Read Response Body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", errors.Wrap(err, "fail to read response body")
	}

	keySignResp := struct {
		R      string `json:"r"`
		S      string `json:"s"`
		Status int    `json:"status"`
	}{}

	if err := json.Unmarshal(respBody, &keySignResp); nil != err {
		return "", "", errors.Wrap(err, "fail to unmarshal tss response body")
	}
	return keySignResp.R, keySignResp.S, nil

}
