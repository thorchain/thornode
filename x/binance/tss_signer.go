package binance

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/btcsuite/btcd/btcec"
	ctypes "github.com/cbarraford/go-sdk/common/types"
	"github.com/cbarraford/go-sdk/common/uuid"
	"github.com/cbarraford/go-sdk/keys"
	"github.com/cbarraford/go-sdk/types/tx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"gitlab.com/thorchain/bepswap/thornode/config"
)

// TSSSigner is a proxy between signer and TSS
type TSSSigner struct {
	cfg    config.TSSConfiguration
	logger zerolog.Logger
	client *http.Client
	addr   ctypes.AccAddress
}

// NewTSSSigner create a new instance of TSSSigner
func NewTSSSigner(cfg config.TSSConfiguration, addr string) (*TSSSigner, error) {
	if len(addr) == 0 {
		return nil, errors.New("tss address is empty")
	}
	if len(cfg.Host) == 0 {
		return nil, errors.New("TSS host is empty")
	}
	if cfg.Port == 0 {
		return nil, errors.New("TSS port not specified")
	}
	accountAddr, err := ctypes.AccAddressFromBech32(addr)
	if nil != err {
		return nil, errors.Wrap(err, "invalid tss account address")
	}
	return &TSSSigner{
		cfg:    cfg,
		logger: log.With().Str("module", "tss_signer").Logger(),
		client: &http.Client{
			Timeout: time.Second * 30,
		},
		addr: accountAddr,
	}, nil
}

// GetPrivKey we don't actually have any private key , but just return something
func (s *TSSSigner) GetPrivKey() crypto.PrivKey {
	return nil
}

func (s *TSSSigner) GetAddr() ctypes.AccAddress {
	return s.addr
}

// ExportAsMnemonic we don't need this function for TSS, just keep it to fulfill KeyManager interface
func (s *TSSSigner) ExportAsMnemonic() (string, error) {
	return "", nil
}

// ExportAsPrivateKey we don't need this function for TSS, just keep it to fulfill KeyManager interface
func (s *TSSSigner) ExportAsPrivateKey() (string, error) {
	return "", nil
}

// ExportAsKeyStore we don't need this function for TSS, just keep it to fulfill KeyManager interface
func (s *TSSSigner) ExportAsKeyStore(password string) (*keys.EncryptedKeyJSON, error) {
	return nil, nil
}
func (s *TSSSigner) makeSignature(msg tx.StdSignMsg) (sig tx.StdSignature, err error) {
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

func (s *TSSSigner) Sign(msg tx.StdSignMsg) ([]byte, error) {
	sig, err := s.makeSignature(msg)
	if err != nil {
		return nil, err
	}
	newTx := tx.NewStdTx(msg.Msgs, []tx.StdSignature{sig}, msg.Memo, msg.Source, msg.Data)
	bz, err := tx.Cdc.MarshalBinaryLengthPrefixed(&newTx)
	if err != nil {
		return nil, err
	}
	// return bz, nil
	return []byte(hex.EncodeToString(bz)), nil
}

func (s *TSSSigner) remoteSign(msg []byte) (SignPack, error) {
	var signPack SignPack
	if len(msg) == 0 {
		return signPack, nil
	}
	channel, err := uuid.NewV4()
	if nil != err {
		return signPack, errors.Wrap(err, "fail to create a new uuid")
	}

	encodedMsg := base64.StdEncoding.EncodeToString(msg)
	signature, err := s.toLocalTSSSigner(channel.String(), encodedMsg)
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
func (s *TSSSigner) getTSSLocalUrl() string {
	u := url.URL{
		Scheme: s.cfg.Scheme,
		Host:   fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port),
	}
	return u.String()
}

// toLocalTSSSigner will send the request to local signer
func (s *TSSSigner) toLocalTSSSigner(nodeid, sendmsg string) (string, error) {
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
