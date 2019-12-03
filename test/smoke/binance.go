package smoke

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	ttypes "github.com/binance-chain/go-sdk/types"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/types/tx"
	"github.com/pkg/errors"
	btypes "gitlab.com/thorchain/thornode/bifrost/binance/types"
)

type Binance struct {
	debug   bool
	host    string
	chainId ctypes.ChainNetwork
}

// NewBinance : new instnance of Binance.
func NewBinance(host string, debug bool) Binance {

	if !strings.HasPrefix(host, "http") {
		host = fmt.Sprintf("http://%s", host)
	}

	ctypes.Network = ctypes.TestNetwork
	return Binance{
		debug:   debug,
		host:    host,
		chainId: ctypes.TestNetwork,
	}
}

func (b Binance) GetBlockHeight() (int64, error) {
	u, err := url.Parse(b.host)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse url")
	}
	u.Path = "block"

	resp, err := http.Get(u.String())
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var tx btypes.RPCBlock
	if err := json.Unmarshal(body, &tx); nil != err {
		return 0, errors.Wrap(err, "fail to unmarshal body")
	}
	block := tx.Result.Block.Header.Height

	parsedBlock, err := strconv.ParseInt(block, 10, 64)
	if nil != err {
		return 0, errors.Wrap(err, "fail to convert block height to int")
	}
	return parsedBlock, nil
}

func (b Binance) GetAccount(addr ctypes.AccAddress) (ctypes.BaseAccount, error) {
	u, err := url.Parse(b.host)
	if err != nil {
		return ctypes.BaseAccount{}, errors.Wrap(err, "failed to parse url")
	}
	u.Path = "abci_query"
	u.RawQuery = fmt.Sprintf("path=\"/account/%s\"", addr.String())

	resp, err := http.Get(u.String())
	if err != nil {
		return ctypes.BaseAccount{}, err
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
		return ctypes.BaseAccount{}, err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return ctypes.BaseAccount{}, err
	}

	data, err := base64.StdEncoding.DecodeString(result.Result.Response.Value)
	if err != nil {
		return ctypes.BaseAccount{}, err
	}

	cdc := ttypes.NewCodec()
	var acc ctypes.AppAccount
	err = cdc.UnmarshalBinaryBare(data, &acc)

	return acc.BaseAccount, err
}

// Input : Prep our input message.
func (b Binance) Input(addr ctypes.AccAddress, coins ctypes.Coins) msg.Input {
	input := msg.Input{
		Address: addr,
		Coins:   coins,
	}

	return input
}

// Output : Prep our output message.
func (b Binance) Output(addr ctypes.AccAddress, coins ctypes.Coins) msg.Output {
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
func (b Binance) CreateMsg(from ctypes.AccAddress, fromCoins ctypes.Coins, transfers []msg.Transfer) msg.SendMsg {
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
	fromCoins := ctypes.Coins{}
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

	signMsg := tx.StdSignMsg{
		ChainID:       "Binance-Chain-Nile", // smoke tests always run on testnet
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
		return nil, nil, fmt.Errorf("No signature returned")
	}
	hexTx := []byte(hex.EncodeToString(rawBz))
	return hexTx, param, nil
}

func (b *Binance) BroadcastTx(hexTx []byte, param map[string]string) error {

	u, err := url.Parse(b.host)
	if err != nil {
		log.Fatal(err)
	}
	u.Path = "broadcast_tx_commit"

	if param != nil {
		q := url.Values{}
		for key, value := range param {
			q.Set(key, value)
		}
		if query := q.Encode(); len(query) > 0 {
			u.RawQuery = query
		}
	}

	resp, err := http.Post(u.String(), "", bytes.NewReader(hexTx))
	if err != nil {
		return errors.Wrap(err, "fail to broadcast tx to ")
	}
	if b.debug {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Printf("Broadcast: %s\n", body)
	}

	return nil
}
