package tss

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	"gitlab.com/thorchain/thornode/common"
)

// KeyGen is
type KeyGen struct {
	keys      *thorclient.Keys
	keyGenCfg config.TSSConfiguration
	logger    zerolog.Logger
	client    *http.Client
}

// NewTssKeyGen create a new instance of TssKeyGen which will look after TSS key stuff
func NewTssKeyGen(keyGenCfg config.TSSConfiguration, keys *thorclient.Keys) (*KeyGen, error) {
	if nil == keys {
		return nil, fmt.Errorf("keys is nil")
	}
	return &KeyGen{
		keys:      keys,
		keyGenCfg: keyGenCfg,
		logger:    log.With().Str("module", "tss_keygen").Logger(),
		client: &http.Client{
			Timeout: time.Second * 130,
		},
	}, nil
}

func (kg *KeyGen) GenerateNewKey(pKeys common.PubKeys) (common.PubKeySet, error) {
	// No need to do key gen
	if len(pKeys) == 0 {
		return common.EmptyPubKeySet, nil
	}
	var keys []string
	for _, item := range pKeys {
		keys = append(keys, item.String())
	}
	keyGenReq := KeyGenRequest{
		Keys: keys,
	}
	buf, err := json.Marshal(keyGenReq)
	if nil != err {
		return common.EmptyPubKeySet, fmt.Errorf("fail to marshal key gen request to json,err:%w", err)
	}
	tssUrl := kg.getTSSLocalUrl()
	kg.logger.Debug().Str("url", tssUrl).Msg("sending request to tss key gen")
	resp, err := kg.client.Post(tssUrl, "application/json", bytes.NewBuffer(buf))
	if nil != err {
		return common.EmptyPubKeySet, fmt.Errorf("fail to send key gen request,err:%w", err)
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			kg.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return common.EmptyPubKeySet, fmt.Errorf("status code from tss keygen (%d)", resp.StatusCode)
	}
	bodyBuf, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return common.EmptyPubKeySet, fmt.Errorf("fail to read response body,err:%w", err)
	}
	var dat KeyGenResp
	err = json.Unmarshal(bodyBuf, &dat)
	if err != nil {
		return common.EmptyPubKeySet, fmt.Errorf("fail to unmarshal tss keygen response,err:%w", err)
	}
	cpk, err := common.NewPubKey(dat.PubKey)
	if nil != err {
		return common.EmptyPubKeySet, fmt.Errorf("fail to create common.PubKey,%w", err)
	}

	// TODO later on THORNode need to have both secp256k1 key and ed25519
	return common.NewPubKeySet(cpk, cpk), nil
}

func (kg *KeyGen) getTSSLocalUrl() string {
	u := url.URL{
		Scheme: kg.keyGenCfg.Scheme,
		Host:   fmt.Sprintf("%s:%d", kg.keyGenCfg.Host, kg.keyGenCfg.Port),
		Path:   "keygen",
	}
	return u.String()
}
