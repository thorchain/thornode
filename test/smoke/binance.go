package smoke

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	sdk "github.com/binance-chain/go-sdk/client"
	"github.com/binance-chain/go-sdk/client/basic"
	"github.com/binance-chain/go-sdk/client/query"
	btypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/types/tx"

	"github.com/go-resty/resty/v2"
)

type Binance struct {
	debug   bool
	delay   time.Duration
	apiHost string
	chainId string
	bClient basic.BasicClient
	qClient query.QueryClient
}

// NewBinance : new instnance of Binance.
func NewBinance(apiHost, chainId string, debug bool) Binance {
	bClient := basic.NewClient(apiHost)
	return Binance{
		debug:   debug,
		delay:   2 * time.Second,
		apiHost: apiHost,
		chainId: chainId,
		bClient: bClient,
		qClient: query.NewClient(bClient),
	}
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
func (b Binance) ParseTx(key keys.KeyManager, transfers []msg.Transfer) msg.SendMsg {
	fromAddr := key.GetAddr()
	fromCoins := btypes.Coins{}
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		fromCoins = fromCoins.Plus(t.Coins)
	}

	sendMsg := b.CreateMsg(fromAddr, fromCoins, transfers)
	return sendMsg
}

// SendTxn : prep and broadcast the transaction to Binance.
func (b Binance) SendTxn(client sdk.DexClient, key keys.KeyManager, payload []msg.Transfer, memo string) {	
	time.Sleep(b.delay)

	if b.debug == true {
		log.Printf("\tFrom: %v", key.GetAddr().String())
		log.Printf("\tMemo: %v\n", memo)
		log.Printf("\tPayload for Binance: %v\n", payload)
	}

	sendMsg := b.ParseTx(key, payload)

	acc, err := b.qClient.GetAccount(key.GetAddr().String())
	if err != nil {
		log.Printf("Error: %v", err)
	}

	signMsg := &tx.StdSignMsg{
		ChainID:       b.chainId,
		Memo:          memo,
		Msgs:          []msg.Msg{sendMsg},
		Source:        tx.Source,
		Sequence:      acc.Sequence,
		AccountNumber: acc.Number,
	}

	rawBz, err := key.Sign(*signMsg)
	if nil != err {
		log.Fatalf("%v", err)
	}
	hexTx := []byte(hex.EncodeToString(rawBz))
	param := map[string]string{
		"sync": "true",
	}

	uri := fmt.Sprintf("%s://%s%s/broadcast",
		types.DefaultApiSchema,
		b.apiHost,
		types.DefaultAPIVersionPrefix)

	rclient := resty.New()

	for i := 1;  i<=5; i++ {
	 

	resp, err := rclient.R().
		SetHeader("Content-Type", "text/plain").
		SetBody(hexTx).
		SetQueryParams(param).
		Post(uri)

	if err != nil {
		log.Printf("%v\n", err);
		time.Sleep(1 * time.Second)

	}else {
		break
    }

	if b.debug == true {
		log.Printf("Commit Response from Binance: %v", string(resp.Body()))
	}
}
}
