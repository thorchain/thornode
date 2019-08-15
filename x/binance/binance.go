package binance

import (
	"os"
	"strconv"

	log "github.com/rs/zerolog/log"

	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/tx"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/client/query"
	"github.com/binance-chain/go-sdk/client/basic"
	sdk "github.com/binance-chain/go-sdk/client"

	ctypes "gitlab.com/thorchain/bepswap/observe/common/types"
	btypes "gitlab.com/thorchain/bepswap/observe/x/binance/types" 
	stypes "gitlab.com/thorchain/bepswap/observe/x/statechain/types" 
)

type Binance struct {
	Client sdk.DexClient
	BasicClient basic.BasicClient
	QueryClient query.QueryClient
	KeyManager keys.KeyManager
	chainId string
}

func NewBinance() *Binance {
	if ctypes.PrivKey == "" {
		log.Fatal().Msgf("No private key set!")
		os.Exit(1)
	}

	keyManager, err := keys.NewPrivateKeyManager(ctypes.PrivKey)
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
		os.Exit(1)
	}

	bClient, err := sdk.NewDexClient(ctypes.DEXHost, types.TestNetwork, keyManager)
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
		os.Exit(1)
	}

	basicClient := basic.NewClient(ctypes.DEXHost)
	queryClient := query.NewClient(basicClient)

	return &Binance{
		Client: bClient,
		BasicClient: basicClient,
		QueryClient: queryClient,
		KeyManager: keyManager,
		// @todo Get this from the transaction client
		chainId: "Binance-Chain-Nile",
	}
}

func (b Binance) Input(addr types.AccAddress, coins types.Coins) msg.Input {
	input := msg.Input{
		Address: addr,
		Coins:   coins,
	}

	return input
}

func (b Binance) Output(addr types.AccAddress, coins types.Coins) msg.Output {
	output := msg.Output{
		Address: addr,
		Coins:   coins,
	}

	return output
}

func (b Binance) MsgToSend(in []msg.Input, out []msg.Output) msg.SendMsg {
	return msg.SendMsg{Inputs: in, Outputs: out}
}

func (b Binance) CreateMsg(from types.AccAddress, fromCoins types.Coins, transfers []msg.Transfer) msg.SendMsg {
	input := b.Input(from, fromCoins)

	output := make([]msg.Output, 0, len(transfers))
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		output = append(output, b.Output(t.ToAddr, t.Coins))
	}

	msg := b.MsgToSend([]msg.Input{input}, output)
	return msg
}

func (b Binance) ParseTx(transfers []msg.Transfer) msg.SendMsg {
	fromAddr := b.KeyManager.GetAddr()
	fromCoins := types.Coins{}
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		fromCoins = fromCoins.Plus(t.Coins)
	}

	sendMsg := b.CreateMsg(fromAddr, fromCoins, transfers)
	return sendMsg
}

func (b Binance) SignTx(txOut stypes.TxOut) ([]byte, map[string]string) {
	var payload []msg.Transfer
	for _, txn := range txOut.TxArray {
		toAddr, _ := types.AccAddressFromBech32(string(types.AccAddress(txn.To)))
		for _, coin := range txn.Coins {
			amount, _ := strconv.ParseInt(coin.Amount, 10, 64)
			payload = append(payload, msg.Transfer{
				toAddr,
				types.Coins{
					types.Coin{
						Denom: coin.Denom,
						Amount: amount,
					},
				},
			})
		}
	}

	sendMsg := b.ParseTx(payload)

	fromAddr := b.KeyManager.GetAddr()
	acc, err := b.QueryClient.GetAccount(fromAddr.String())
	if err != nil {
		log.Error().Msgf("Error: %v", err)
	}

	signMsg := &tx.StdSignMsg{
		ChainID: 				b.chainId,
		Memo:    				btypes.TxOutMemoPrefix+txOut.Height,
		Msgs:    				[]msg.Msg{sendMsg},
		Source:  				tx.Source,
		Sequence: 			acc.Sequence,
		AccountNumber: 	acc.Number,
	}

	hexTx, _ := b.KeyManager.Sign(*signMsg)
	param := map[string]string{}
	param["sync"] = "true"

	return hexTx, param
}

func (b Binance) BroadcastTx(hexTx []byte, param map[string]string) (*tx.TxCommitResult, error) {
	commits, err := b.Client.PostTx(hexTx, param)
	if err != nil {
		log.Error().Msgf("Error: %v", err)
		return nil, err
	}

	log.Info().Msgf("Commit Response from Binance: %v", commits[0])
	return &commits[0], nil
}
