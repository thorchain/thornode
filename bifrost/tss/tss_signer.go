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
	"github.com/btcsuite/btcd/btcec"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"gitlab.com/thorchain/thornode/bifrost/config"

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

// GetPrivKey we don't actually have any private key , but just return something
func (s *KeySign) GetPrivKey() crypto.PrivKey {
	return nil
}

func (s *KeySign) GetAddr() ctypes.AccAddress {
	return nil
}

// ExportAsMnemonic we don't need this function for TSS, just keep it to fulfill KeyManager interface
func (s *KeySign) ExportAsMnemonic() (string, error) {
	return "", nil
}

// ExportAsPrivateKey we don't need this function for TSS, just keep it to fulfill KeyManager interface
func (s *KeySign) ExportAsPrivateKey() (string, error) {
	return "", nil
}

// ExportAsKeyStore we don't need this function for TSS, just keep it to fulfill KeyManager interface
func (s *KeySign) ExportAsKeyStore(password string) (*keys.EncryptedKeyJSON, error) {
	return nil, nil
}
func (s *KeySign) makeSignature(msg tx.StdSignMsg) (sig tx.StdSignature, err error) {
	var stdSignature tx.StdSignature
	signPack, err := s.remoteSign(msg.Bytes())
	if err != nil {
		return stdSignature, errors.Wrap(err, "fail to TSS sign")
	}
	R, _ := new(big.Int).SetString(signPack.R, 10)
	S, _ := new(big.Int).SetString(signPack.S, 10)

	N := btcec.S256().N
	halfOrder := new(big.Int).Rsh(N, 1)

	Pubx, _ := new(big.Int).SetString(signPack.Pubkeyx, 10)
	Puby, _ := new(big.Int).SetString(signPack.Pubkeyy, 10)

	// see: https://github.com/ethereum/go-ethereum/blob/
	// f9401ae011ddf7f8d2d95020b7446c17f8d98dc1/crypto/signature_nocgo.go#L90-L93
	if S.Cmp(halfOrder) == 1 {
		S.Sub(N, S)
	}

	// Serialize signature to R || S.
	// R, S are padded to 32 bytes respectively.
	rBytes := R.Bytes()
	sBytes := S.Bytes()
	sigBytes := make([]byte, 64)
	// 0 pad the byte arrays from the left if they aren't big enough.
	copy(sigBytes[32-len(rBytes):32], rBytes)
	copy(sigBytes[64-len(sBytes):64], sBytes)

	tsspubkey := btcec.PublicKey{
		Curve: btcec.S256(),
		X:     Pubx,
		Y:     Puby,
	}

	var pubkeyBytes secp256k1.PubKeySecp256k1
	copy(pubkeyBytes[:], tsspubkey.SerializeCompressed())

	return tx.StdSignature{
		AccountNumber: msg.AccountNumber,
		Sequence:      msg.Sequence,
		PubKey:        pubkeyBytes,
		Signature:     sigBytes,
	}, nil
}

func (s *KeySign) Sign(msg tx.StdSignMsg) ([]byte, error) {
	sig, err := s.makeSignature(msg)
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

func (s *KeySign) remoteSign(msg []byte) (SignPack, error) {
	var signPack SignPack
	if len(msg) == 0 {
		return signPack, nil
	}
	encodedMsg := base64.StdEncoding.EncodeToString(msg)
	signature, err := s.toLocalTSSSigner(s.cfg.NodeId, encodedMsg)
	if nil != err {
		return signPack, errors.Wrap(err, "fail to tss sign")
	}

	if signature == "BROKEN SIGNATURE" {
		return signPack, errors.New("BROKEN SIGNATURE")
	}
	s.logger.Debug().Str("signature", signature).Msg("tss result")
	data, err := base64.StdEncoding.DecodeString(signature)
	if nil != err {
		return signPack, errors.Wrap(err, "fail to decode tss signature")
	}

	if err := json.Unmarshal(data, &signPack); nil != err {
		return signPack, errors.Wrap(err, "fail to unmarshal result to signPack")
	}
	return signPack, nil
}
func (s *KeySign) getTSSLocalUrl() string {
	u := url.URL{
		Scheme: s.cfg.Scheme,
		Host:   fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port),
		Path:   "recvmsg",
	}
	return u.String()
}

// toLocalTSSSigner will send the request to local signer
func (s *KeySign) toLocalTSSSigner(nodeid, sendmsg string) (string, error) {
	tssMsg := struct {
		NodeID string `json:"Nodeid"`
		Msg    string `json:"Msg"`
	}{
		NodeID: nodeid,
		Msg:    sendmsg,
	}
	buf, err := json.Marshal(tssMsg)
	if nil != err {
		return "", errors.Wrap(err, "fail to create tss request msg")
	}
	s.logger.Debug().Str("payload", string(buf)).Msg("msg to tss Local node")
	localTssURL := s.getTSSLocalUrl()
	resp, err := s.client.Post(localTssURL, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return "", errors.Wrapf(err, "fail to send request to local TSS node,url: %s", localTssURL)
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			s.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("response status: %s from tss sign ", resp.Status)
	}

	// Read Response Body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "fail to read response body")
	}

	// TODO need to double check what's the response and unmarshal it appropriately
	var dat map[string]interface{}
	if err := json.Unmarshal(respBody, &dat); nil != err {
		return "", errors.Wrap(err, "fail to unmarshal tss response body")
	}

	msg := dat["Ok"].(map[string]interface{})
	signature := msg["Msg"].(string)

	return signature, nil

}
