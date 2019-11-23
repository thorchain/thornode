package smoke

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	btypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	ttypes "github.com/binance-chain/go-sdk/types"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/types/tx"
	"github.com/pkg/errors"
)

type Binance struct {
	debug   bool
	delay   time.Duration
	apiHost string
	chainId btypes.ChainNetwork
}

// NewBinance : new instnance of Binance.
func NewBinance(apiHost string, chainId btypes.ChainNetwork, debug bool) Binance {
	btypes.Network = chainId
	return Binance{
		debug:   debug,
		delay:   2 * time.Second,
		apiHost: apiHost,
		chainId: chainId,
	}
}

func (b Binance) GetBalances(address btypes.AccAddress) (btypes.Coins, error) {
	key := append([]byte("account:"), address.Bytes()...)
	args := fmt.Sprintf("path=\"/store/acc/key\"&data=0x%x", key)
	uri := url.URL{
		Scheme:   "http", // TODO: don't hard code this
		Host:     b.apiHost,
		Path:     "abci_query",
		RawQuery: args,
	}
	resp, err := http.Get(uri.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type queryResult struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      string `json:"id"`
		Result  struct {
			Response struct {
				Key         string `json:"key"`
				Value       string `json:"value"`
				BlockHeight string `json:"height"`
			} `json:"response"`
		} `json:"result"`
	}

	var result queryResult
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(result.Result.Response.Value)
	if err != nil {
		return nil, err
	}

	cdc := ttypes.NewCodec()
	var acc btypes.AppAccount
	err = cdc.UnmarshalBinaryBare(data, &acc)

	return acc.BaseAccount.Coins, err
}

func (b Binance) GetAccount(addr btypes.AccAddress) (btypes.BaseAccount, error) {
	path := fmt.Sprintf("/abci_query?path=\"/account/%s\"", addr.String())
	uri := fmt.Sprintf("http://%s%s", b.apiHost, path)
	// TODO: don't hard code to http protocol
	resp, err := http.Get(uri)
	if err != nil {
		return btypes.BaseAccount{}, err
	}
	defer resp.Body.Close()

	type queryResult struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      string `json:"id"`
		Result  struct {
			Response struct {
				Key         string `json:"key"`
				Value       string `json:"value"`
				BlockHeight string `json:"height"`
			} `json:"response"`
		} `json:"result"`
	}

	var result queryResult
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return btypes.BaseAccount{}, err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return btypes.BaseAccount{}, err
	}

	data, err := base64.StdEncoding.DecodeString(result.Result.Response.Value)
	if err != nil {
		return btypes.BaseAccount{}, err
	}

	cdc := ttypes.NewCodec()
	var acc btypes.BaseAccount
	err = cdc.UnmarshalBinaryBare(data, &acc)

	return acc, err
}

// Input : Prep our input message.
func (b Binance) Input(addr btypes.AccAddress, coins btypes.Coins) msg.Input {
	input := msg.Input{
		Address: addr,
		Coins:   coins,
	}

	return input
}

// Output : Prep our output message.
func (b Binance) Output(addr btypes.AccAddress, coins btypes.Coins) msg.Output {
	output := msg.Output{
		Address: addr,
		Coins:   coins,
	}

	return output
}

// MsgToSend : Prep the message to send.
func (b Binance) MsgToSend(in []msg.Input, out []msg.Output) msg.SendMsg {
	return msg.SendMsg{Inputs: in, Outputs: out}
}

// CreateMsg : Create a new message to broadcast to Binance.
func (b Binance) CreateMsg(from btypes.AccAddress, fromCoins btypes.Coins, transfers []msg.Transfer) msg.SendMsg {
	input := b.Input(from, fromCoins)

	output := make([]msg.Output, 0, len(transfers))
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		output = append(output, b.Output(t.ToAddr, t.Coins))
	}

	msg := b.MsgToSend([]msg.Input{input}, output)
	return msg
}

// ParseTx : Parse the transaction.
func (b Binance) ParseTx(key keys.KeyManager, transfers []msg.Transfer) (msg.SendMsg, error) {
	fromAddr := key.GetAddr()
	fromCoins := btypes.Coins{}
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		fromCoins = fromCoins.Plus(t.Coins)
	}

	sendMsg := b.CreateMsg(fromAddr, fromCoins, transfers)
	err := sendMsg.ValidateBasic()
	return sendMsg, err
}

func (b Binance) SignTx(key keys.KeyManager, sendMsg msg.SendMsg, memo string) ([]byte, map[string]string, error) {
	acc, err := b.GetAccount(key.GetAddr())
	if err != nil {
		return nil, nil, errors.Wrap(err, "fail to get account info")
	}

	chainId := "Binance-Chain-Tigris"
	if b.chainId == btypes.TestNetwork {
		chainId = "Binance-Chain-Nile"
	}

	signMsg := tx.StdSignMsg{
		ChainID:       chainId,
		Memo:          memo,
		Msgs:          []msg.Msg{sendMsg},
		Source:        tx.Source,
		Sequence:      acc.GetSequence(),
		AccountNumber: acc.GetAccountNumber(),
	}
	param := map[string]string{
		"sync": "true",
	}
	rawBz, err := key.Sign(signMsg)
	if nil != err {
		return nil, nil, errors.Wrap(err, "fail to sign message")
	}

	if len(rawBz) == 0 {
		return nil, nil, nil
	}
	hexTx := []byte(hex.EncodeToString(rawBz))
	return hexTx, param, nil
}

func (b *Binance) BroadcastTx(hexTx []byte, param map[string]string) (*tx.TxCommitResult, error) {
	uri := url.URL{
		Scheme: "http", // TODO: don't hard code this
		Host:   b.apiHost,
		Path:   "broadcast_tx_commit",
		// TODO: add params as query args

	}
	resp, err := http.Post(uri.String(), "", bytes.NewReader(hexTx))
	if err != nil {
		return nil, errors.Wrap(err, "fail to broadcast tx to ")
	}
	if b.debug {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Broadcast: %s\n", body)
	}
	return nil, nil
}
