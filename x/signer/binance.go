package signer

import (
	"os"
	"strconv"
	log "github.com/rs/zerolog/log"

	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/tx"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/common/types"
	sdk "github.com/binance-chain/go-sdk/client"

	stypes "gitlab.com/thorchain/bepswap/observe/x/signer/types"
)

type Binance struct {
	PoolAddress string
	PrivateKey string
	DexHost string
	Client sdk.DexClient
	KeyManager keys.KeyManager
	chainId string
}

func NewBinance(poolAddress, dexHost string) *Binance {
	key := os.Getenv("PRIVATE_KEY")
	if key == "" {
		log.Fatal().Msg("No private key set!")
		os.Exit(1)
	}

	keyManager, err := keys.NewPrivateKeyManager(key)
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
		os.Exit(1)
	}

	bClient, err := sdk.NewDexClient(dexHost, types.TestNetwork, keyManager)
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
		os.Exit(1)
	}

	return &Binance{
		PrivateKey: key,
		DexHost: dexHost,
		Client: bClient,
		KeyManager: keyManager,
		// @todo Get this from the transaction client
		chainId: "Binance-Chain-Nile",
	}
}

func (b *Binance) Input(addr types.AccAddress, coins types.Coins) msg.Input {
	input := msg.Input{
		Address: addr,
		Coins:   coins,
	}

	return input
}

func (b *Binance) Output(addr types.AccAddress, coins types.Coins) msg.Output {
	output := msg.Output{
		Address: addr,
		Coins:   coins,
	}

	return output
}

func (b *Binance) MsgToSend(in []msg.Input, out []msg.Output) msg.SendMsg {
	return msg.SendMsg{Inputs: in, Outputs: out}
}

func (b *Binance) CreateMsg(from types.AccAddress, fromCoins types.Coins, transfers []msg.Transfer) msg.SendMsg {
	input := b.Input(from, fromCoins)

	output := make([]msg.Output, 0, len(transfers))
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		output = append(output, b.Output(t.ToAddr, t.Coins))
	}

	msg := b.MsgToSend([]msg.Input{input}, output)
	return msg
}

func (b *Binance) ParseTx(transfers []msg.Transfer) msg.SendMsg {
	fromAddr := b.KeyManager.GetAddr()
	fromCoins := types.Coins{}
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		fromCoins = fromCoins.Plus(t.Coins)
	}

	sendMsg := b.CreateMsg(fromAddr, fromCoins, transfers)
	return sendMsg
}

func (b *Binance) SignTx(outTx stypes.OutTx) ([]byte, map[string]string) {
	//var options tx.StdSignMsg
	//options.Memo = outTx.TxOutID

	var payload []msg.Transfer
	for _, txn := range outTx.TxArray {
		toAddr, _ := types.AccAddressFromBech32(string(types.AccAddress(txn.To)))
		for _, coin := range txn.Coins {
			parsedAmt, _ := strconv.ParseInt(coin.Amount, 10, 64)
			amount := parsedAmt*100000000

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
	signMsg := &tx.StdSignMsg{
		ChainID: b.chainId,
		Memo:    "", //outTx.TxOutID,
		Msgs:    []msg.Msg{sendMsg},
		Source:  tx.Source,
	}

	hexTx, _ := b.KeyManager.Sign(*signMsg)
	param := map[string]string{}
	param["sync"] = "true"

	return hexTx, param
}

func (b *Binance) BroadcastTx(hexTx []byte, param map[string]string) *tx.TxCommitResult {
	commits, _ := b.Client.PostTx(hexTx, param)
	return &commits[0]
}

