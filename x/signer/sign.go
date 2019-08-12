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

type Signer struct {
	PoolAddress string
	PrivateKey string
	DexHost string
	Client sdk.DexClient
	KeyManager keys.KeyManager
	chainId string
}

func NewSigner(poolAddress, dexHost string) *Signer {
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

	return &Signer{
		PrivateKey: key,
		DexHost: dexHost,
		Client: bClient,
		KeyManager: keyManager,
		// @todo Get this from the transaction client
		chainId: "Binance-Chain-Nile",
	}
}

func (s *Signer) Input(addr types.AccAddress, coins types.Coins) msg.Input {
	input := msg.Input{
		Address: addr,
		Coins:   coins,
	}

	return input
}

func (s *Signer) Output(addr types.AccAddress, coins types.Coins) msg.Output {
	output := msg.Output{
		Address: addr,
		Coins:   coins,
	}

	return output
}

func (s *Signer) MsgToSend(in []msg.Input, out []msg.Output) msg.SendMsg {
	return msg.SendMsg{Inputs: in, Outputs: out}
}

func (s *Signer) CreateMsg(from types.AccAddress, fromCoins types.Coins, transfers []msg.Transfer) msg.SendMsg {
	input := s.Input(from, fromCoins)

	output := make([]msg.Output, 0, len(transfers))
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		output = append(output, s.Output(t.ToAddr, t.Coins))
	}

	msg := s.MsgToSend([]msg.Input{input}, output)
	return msg
}

func (s *Signer) ParseTx(transfers []msg.Transfer) msg.SendMsg {
	fromAddr := s.KeyManager.GetAddr()
	fromCoins := types.Coins{}
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		fromCoins = fromCoins.Plus(t.Coins)
	}

	sendMsg := s.CreateMsg(fromAddr, fromCoins, transfers)
	return sendMsg
}

func (s *Signer) SignTx(outTx stypes.OutTx) ([]byte, map[string]string) {
	var options tx.StdSignMsg
	options.Memo = outTx.TxOutID

	var payload []msg.Transfer
	for _, txn := range outTx.TxArray {
		toAddr, _ := types.AccAddressFromBech32(string(types.AccAddress(txn.To)))
		amount, _ := strconv.ParseInt(txn.Amount, 10, 64)
		payload = append(payload, msg.Transfer{toAddr, types.Coins{types.Coin{Denom: txn.Token, Amount: amount}}})
	}

	sendMsg := s.ParseTx(payload)
	signMsg := &tx.StdSignMsg{
		ChainID: s.chainId,
		Memo:    outTx.TxOutID,
		Msgs:    []msg.Msg{sendMsg},
		Source:  tx.Source,
	}

	hexTx, _ := s.KeyManager.Sign(*signMsg)
	param := map[string]string{}
	param["sync"] = "true"

	return hexTx, param
}

func (s *Signer) BroadcastTx(hexTx []byte, param map[string]string) *tx.TxCommitResult {
	commits, _ := s.Client.PostTx(hexTx, param)
	return &commits[0]
}
