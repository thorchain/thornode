package exchange

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	c := Client{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}

	return &c
}

func (cli *Client) GetTxInfo(txHash string) (tx txResult, err error) {
	// TODO: support mainnet

	req, err := http.NewRequest("GET", fmt.Sprintf("https://testnet-dex.binance.org/api/v1/tx/%s?format=json", txHash), nil)
	if err != nil {
		return tx, errors.Wrap(err, "failed to build request")
	}

	resp, err := cli.httpClient.Do(req)
	if err != nil {
		return tx, errors.Wrap(err, "request failed")
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&tx); err != nil {
		return tx, errors.Wrap(err, "unmarshaling failed")
	}
	return
}

type msg struct {
	Type  string `json:"type"`
	Value struct {
		Inputs  []puts `json:"inputs"`
		Outputs []puts `json:"outputs"`
	} `json:"value"`
}

type puts struct {
	Address string    `json:"address"`
	Coins   sdk.Coins `json:"coins"`
}

type txResult struct {
	Tx struct {
		Value struct {
			Memo string `json:"memo"`
			Msg  []msg  `json:"msg"`
		} `json:"value"`
	} `json:"tx"`
}

func (tx txResult) Memo() string {
	return tx.Tx.Value.Memo
}

func (tx txResult) Msg() msg {
	msgs := tx.Tx.Value.Msg
	if len(msgs) == 0 {
		return msg{}
	}
	return msgs[0]
}

func (tx txResult) Outputs() []puts {
	return tx.Msg().Value.Outputs
}

func (tx txResult) Inputs() []puts {
	return tx.Msg().Value.Inputs
}
