package signer

import (
	"fmt"
	"net/url"
	"io/ioutil"
	"encoding/json"

	//log "github.com/rs/zerolog/log"
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

func (s *StateChain) Query(blockHeight int64) types.OutTx {
	path := fmt.Sprintf("/swapservice/txoutarray/%v", blockHeight)
	uri := url.URL{
		Scheme: "http",
		Host: s.ChainHost,
		Path: path,
	}

	resp, _ := http.Get(uri.String())
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var outTx types.OutTx
	json.Unmarshal(body, &outTx)

	return outTx
}
