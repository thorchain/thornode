package exchange

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var netClient = &http.Client{
	Timeout: time.Second * 10,
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

func GetTxInfo(txHash string) (tx txResult, err error) {
	response, err := http.Get(fmt.Sprintf("https://testnet-dex.binance.org/api/v1/tx/%s?format=json", txHash))
	if err != nil {
		return
	}
	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	json.Unmarshal(buf, &tx)
	return
}
