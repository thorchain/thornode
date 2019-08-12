package signer

import (
	"fmt"
	"net/url"
	"io/ioutil"
	"encoding/json"

	log "github.com/rs/zerolog/log"
	http "github.com/hashicorp/go-retryablehttp"

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

	log.Info().Msgf("Querying Height from %v", uri.String())

	resp, _ := http.Get(uri.String())
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

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

	resp, _ := http.Get(uri.String())
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var outTx types.OutTx
	err := json.Unmarshal(body, &outTx)
	if err != nil {
		log.Error().Msgf("Error: %v", err)
	}

	return outTx
}
