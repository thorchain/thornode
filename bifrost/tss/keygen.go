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
	keys          *thorclient.Keys
	keyGenCfg     config.TSSConfiguration
	stateChainCfg config.StateChainConfiguration
	logger        zerolog.Logger
	client        *http.Client
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
			Timeout: time.Second * 30,
		},
	}, nil
}

// getValidatorKeys from thorchain
func (kg *KeyGen) getValidatorKeys() ([]common.PubKey, error) {
	resp, err := thorclient.GetValidators(kg.client, kg.stateChainCfg.ChainHost)
	if nil != err {
		return nil, fmt.Errorf("fail to get validators , err:%w", err)
	}
	noNominated := resp.Nominated == nil || resp.Nominated.IsEmpty()
	noQueued := resp.Queued == nil || resp.Queued.IsEmpty()
	if noNominated && noQueued {
		kg.logger.Info().Msg("no node get nominated , and no node get queued to be rotate out, so don't need to rotate poo")
		return nil, nil
	}
	pKeys := make([]common.PubKey, 0, len(resp.ActiveNodes)+1)
	if !resp.Nominated.IsEmpty() {
		for _, item := range resp.Nominated {
			pKeys = append(pKeys, item.NodePubKey.Secp256k1)
		}

	}
	queued := resp.Queued
	for _, item := range resp.ActiveNodes {
		if queued.Contains(item) {
			continue
		}
		pKeys = append(pKeys, item.NodePubKey.Secp256k1)
	}
	return pKeys, nil
}

func (kg *KeyGen) GenerateNewKey() (common.PubKeys, error) {
	pKeys, err := kg.getValidatorKeys()
	if nil != err {
		return common.EmptyPubKeys, fmt.Errorf("fail to get validator keys from thorchain,err:%w", err)
	}

	// No need to do key gen
	if len(pKeys) == 0 {
		return common.EmptyPubKeys, nil
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
		return common.EmptyPubKeys, fmt.Errorf("fail to marshal key gen request to json,err:%w", err)
	}
	tssUrl := kg.getTSSLocalUrl()
	kg.logger.Debug().Str("url", tssUrl).Msg("sending request to tss key gen")
	resp, err := kg.client.Post(tssUrl, "application/json", bytes.NewBuffer(buf))
	if nil != err {
		return common.EmptyPubKeys, fmt.Errorf("fail to send key gen request,err:%w", err)
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			kg.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return common.EmptyPubKeys, fmt.Errorf("status code from tss keygen (%d)", resp.StatusCode)
	}
	bodyBuf, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return common.EmptyPubKeys, fmt.Errorf("fail to read response body,err:%w", err)
	}
	var dat KeyGenResp
	err = json.Unmarshal(bodyBuf, &dat)
	if err != nil {
		return common.EmptyPubKeys, fmt.Errorf("fail to unmarshal tss keygen response,err:%w", err)
	}
	cpk, err := common.NewPubKey(dat.PubKey)
	if nil != err {
		return common.EmptyPubKeys, fmt.Errorf("fail to create common.PubKey,%w", err)
	}

	// TODO later on THORNode need to have both secp256k1 key and ed25519
	return common.NewPubKeys(cpk, cpk), nil
}
func (kg *KeyGen) getTSSLocalUrl() string {
	u := url.URL{
		Scheme: kg.keyGenCfg.Scheme,
		Host:   fmt.Sprintf("%s:%d", kg.keyGenCfg.Host, kg.keyGenCfg.Port),
		Path:   "keygen",
	}
	return u.String()
}
