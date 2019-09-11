package binance

import (
	"strings"

	sdk "github.com/binance-chain/go-sdk/client"
	"github.com/binance-chain/go-sdk/client/basic"
	"github.com/binance-chain/go-sdk/client/query"
	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/types/tx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/thorchain/bepswap/common"

	"gitlab.com/thorchain/bepswap/observe/config"
	btypes "gitlab.com/thorchain/bepswap/observe/x/binance/types"
	stypes "gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

type Binance struct {
	logger      zerolog.Logger
	cfg         config.BinanceConfiguration
	Client      sdk.DexClient
	basicClient basic.BasicClient
	queryClient query.QueryClient
	keyManager  keys.KeyManager
	chainId     string
}

// NewBinance create new instance of binance client
func NewBinance(cfg config.BinanceConfiguration) (*Binance, error) {
	if len(cfg.PrivateKey) == 0 {
		return nil, errors.New("no private key")
	}
	if len(cfg.DEXHost) == 0 {
		return nil, errors.New("dex host is empty, set env DEX_HOST")
	}

	keyManager, err := keys.NewPrivateKeyManager(cfg.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create private key manager")
	}
	chainNetwork := types.TestNetwork
	if !IsTestNet(cfg.DEXHost) {
		chainNetwork = types.ProdNetwork
	}
	bClient, err := sdk.NewDexClient(cfg.DEXHost, chainNetwork, keyManager)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create binance client")
	}

	basicClient := basic.NewClient(cfg.DEXHost)
	queryClient := query.NewClient(basicClient)
	return &Binance{
		logger:      log.With().Str("module", "binance").Logger(),
		cfg:         cfg,
		Client:      bClient,
		basicClient: basicClient,
		queryClient: queryClient,
		keyManager:  keyManager,
		chainId:     "Binance-Chain-Nile",
	}, nil
}

const (
	testNetUrl = "testnet-dex.binance.org"
)

func IsTestNet(dexHost string) bool {
	return strings.Contains(dexHost, testNetUrl) || strings.Contains(dexHost, "127.0.0.1")
}
func (b *Binance) Input(addr types.AccAddress, coins types.Coins) msg.Input {
	return msg.Input{
		Address: addr,
		Coins:   coins,
	}
}

func (b *Binance) output(addr types.AccAddress, coins types.Coins) msg.Output {
	return msg.Output{
		Address: addr,
		Coins:   coins,
	}
}

func (b *Binance) msgToSend(in []msg.Input, out []msg.Output) msg.SendMsg {
	return msg.SendMsg{Inputs: in, Outputs: out}
}

func (b *Binance) createMsg(from types.AccAddress, fromCoins types.Coins, transfers []msg.Transfer) msg.SendMsg {
	input := b.Input(from, fromCoins)
	output := make([]msg.Output, 0, len(transfers))
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		output = append(output, b.output(t.ToAddr, t.Coins))
	}
	return b.msgToSend([]msg.Input{input}, output)
}

func (b *Binance) parseTx(transfers []msg.Transfer) msg.SendMsg {
	fromAddr := b.keyManager.GetAddr()
	fromCoins := types.Coins{}
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		fromCoins = fromCoins.Plus(t.Coins)
	}
	return b.createMsg(fromAddr, fromCoins, transfers)
}

func (b *Binance) SignTx(txOut stypes.TxOut) ([]byte, map[string]string, error) {
	var payload []msg.Transfer
	for _, txn := range txOut.TxArray {
		toAddr, err := types.AccAddressFromBech32(txn.To)
		if nil != err {
			return nil, nil, errors.Wrapf(err, "fail to parse account address(%s)", txn.To)
		}
		for _, coin := range txn.Coins {
			amount := coin.Amount.Float64()
			ticker := coin.Denom
			if common.IsRune(coin.Denom) {
				// TODO need to change it on mainnet
				ticker = common.RuneA1FTicker
			}
			payload = append(payload, msg.Transfer{
				ToAddr: toAddr,
				Coins: types.Coins{
					types.Coin{
						Denom:  ticker.String(),
						Amount: int64(amount),
					},
				},
			})
		}
	}
	if len(payload) == 0 {
		return nil, nil, nil
	}
	fromAddr := b.keyManager.GetAddr()
	sendMsg := b.parseTx(payload)
	acc, err := b.queryClient.GetAccount(fromAddr.String())
	if err != nil {
		return nil, nil, errors.Wrap(err, "fail to get account info")
	}

	signMsg := tx.StdSignMsg{
		ChainID:       b.chainId,
		Memo:          btypes.TxOutMemoPrefix + txOut.Height,
		Msgs:          []msg.Msg{sendMsg},
		Source:        tx.Source,
		Sequence:      acc.Sequence,
		AccountNumber: acc.Number,
	}

	hexTx, err := b.keyManager.Sign(signMsg)
	if nil != err {
		return nil, nil, errors.Wrap(err, "fail to sign message")
	}
	param := map[string]string{
		"sync": "true",
	}
	return hexTx, param, nil
}

func (b *Binance) BroadcastTx(hexTx []byte, param map[string]string) (*tx.TxCommitResult, error) {
	commits, err := b.Client.PostTx(hexTx, param)
	if err != nil {
		return nil, errors.Wrap(err, "fail to broadcast tx to ")
	}
	for _, commitResult := range commits {
		b.logger.Debug().
			Bool("ok", commitResult.Ok).
			Str("log", commitResult.Log).
			Str("hash", commitResult.Hash).
			Int32("code", commitResult.Code).
			Str("data", commitResult.Data).
			Msg("get commit response from binance")
	}
	return &commits[0], nil
}
