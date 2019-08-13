package signer

import (
	"fmt"
	"errors"
	"net/url"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"github.com/avast/retry-go"
	log "github.com/rs/zerolog/log"

	types "gitlab.com/thorchain/bepswap/observe/x/signer/types"
)

type StateChain struct {
	ChainHost string
}

func NewStateChain(chainHost string) *StateChain {
	return &StateChain{
		ChainHost: chainHost,
	}
}

func (s *StateChain) TxnBlockHeight(txn string) string {
	uri := url.URL{
		Scheme: "http",
		Host: s.ChainHost,
		Path: fmt.Sprintf("/txs/%s", txn),
	}

	body, _ := GetWithRetry(uri.String())

	var txs types.Txs
	err := json.Unmarshal(body, &txs)
	if err != nil {
		log.Error().Msgf("Error: %v", err)
	}

	return txs.Height
}

func (s *StateChain) TxOut(blockHeight string) types.OutTx {
	uri := url.URL{
		Scheme: "http",
		Host: s.ChainHost,
		Path: fmt.Sprintf("/swapservice/txoutarray/%s", blockHeight),
	}

	body, _ := GetWithRetry(uri.String())

	var outTx types.OutTx
	err := json.Unmarshal(body, &outTx)
	if err != nil {
		log.Error().Msgf("Error: %v", err)
	}

	return outTx
}

func GetWithRetry(uri string) ([]byte, error) {
	var body []byte

	err := retry.Do(
		func() error {
			resp, err := http.Get(uri)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode == 404 {
				return errors.New("404")
			}

			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			return nil
		},
	)

	return body, err
}
