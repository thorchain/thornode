package binance

import (
	"encoding/hex"
	"strconv"
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

	"gitlab.com/thorchain/bepswap/thornode/bifrost/config"
	stypes "gitlab.com/thorchain/bepswap/thornode/bifrost/statechain/types"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

type Binance struct {
	logger      zerolog.Logger
	cfg         config.BinanceConfiguration
	Client      sdk.DexClient
	basicClient basic.BasicClient
	queryClient query.QueryClient
	keyManager  keys.KeyManager
	chainId     string
	isTestNet   bool
}

// NewBinance create new instance of binance client
func NewBinance(cfg config.BinanceConfiguration) (*Binance, error) {

	if !cfg.UseTSS && len(cfg.PrivateKey) == 0 {
		return nil, errors.New("no private key")
	}
	if len(cfg.DEXHost) == 0 {
		return nil, errors.New("dex host is empty, set env DEX_HOST")
	}
	var km keys.KeyManager
	var err error
	if cfg.UseTSS {
		km, err = NewTSSSigner(cfg.TSS, cfg.TSSAddress)
		if nil != err {
			return nil, errors.Wrap(err, "fail to create tss signer")
		}
	} else {
		km, err = keys.NewPrivateKeyManager(cfg.PrivateKey)
		if err != nil {
			return nil, errors.Wrap(err, "fail to create private key manager")
		}
	}
	chainNetwork := types.TestNetwork
	isTestNet := IsTestNet(cfg.DEXHost)
	if !isTestNet {
		chainNetwork = types.ProdNetwork
	}
	bClient, err := sdk.NewDexClient(cfg.DEXHost, chainNetwork, km)
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
		keyManager:  km,
		isTestNet:   isTestNet,
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
	fromAddr := b.GetAddress()
	addr, err := types.AccAddressFromBech32(fromAddr)
	if nil != err {
		b.logger.Error().Str("address", fromAddr).Err(err).Msg("fail to parse address")
	}
	fromCoins := types.Coins{}
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		fromCoins = fromCoins.Plus(t.Coins)
	}
	return b.createMsg(addr, fromCoins, transfers)
}

// GetAddress return current signer address
func (b *Binance) GetAddress() string {
	return b.keyManager.GetAddr().String()
}
func (b *Binance) isSignerAddressMatch(poolAddr, signerAddr string) bool {
	pubKey, err := common.NewPubKey(poolAddr)
	if nil != err {
		b.logger.Error().Err(err).Msg("fail to create pub key from the pool address")
		return false
	}
	bnbAddress, err := pubKey.GetAddress(common.BNBChain)
	if nil != err {
		b.logger.Error().Err(err).Msg("fail to create bnb address from the pub key")
		return false
	}
	b.logger.Info().Msg(bnbAddress.String())
	return strings.EqualFold(bnbAddress.String(), signerAddr)
}

func (b *Binance) SignTx(tai stypes.TxArrayItem, height int64) ([]byte, map[string]string, error) {
	signerAddr := b.GetAddress()
	var payload []msg.Transfer

	if !b.isSignerAddressMatch(tai.PoolAddress.String(), signerAddr) {
		b.logger.Info().Str("signer addr", signerAddr).Str("pool addr", tai.PoolAddress.String()).Msg("address doesn't match ignore")
		return nil, nil, nil
	}
	toAddr, err := types.AccAddressFromBech32(tai.To)
	if nil != err {
		return nil, nil, errors.Wrapf(err, "fail to parse account address(%s)", tai.To)
	}
	seqNo, err := strconv.ParseInt(tai.SeqNo, 10, 64)
	if nil != err {
		return nil, nil, errors.Wrapf(err, "fail to parse seq no %s", tai.SeqNo)
	}
	for _, coin := range tai.Coins {
		amount := coin.Amount
		asset := coin.Asset
		if common.IsRuneAsset(coin.Asset) {
			asset = common.RuneAsset()
		}
		payload = append(payload, msg.Transfer{
			ToAddr: toAddr,
			Coins: types.Coins{
				types.Coin{
					Denom:  asset.Symbol.String(),
					Amount: int64(amount.Uint64()),
				},
			},
		})
	}

	if len(payload) == 0 {
		return nil, nil, nil
	}
	fromAddr := b.GetAddress()
	sendMsg := b.parseTx(payload)
	if err := sendMsg.ValidateBasic(); nil != err {
		return nil, nil, errors.Wrap(err, "invalid send msg")
	}
	acc, err := b.queryClient.GetAccount(fromAddr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "fail to get account info")
	}

	signMsg := tx.StdSignMsg{
		ChainID:       b.chainId,
		Memo:          tai.Memo,
		Msgs:          []msg.Msg{sendMsg},
		Source:        tx.Source,
		Sequence:      seqNo, // acc.Sequence,
		AccountNumber: acc.Number,
	}
	param := map[string]string{
		"sync": "true",
	}
	rawBz, err := b.keyManager.Sign(signMsg)
	if nil != err {
		return nil, nil, errors.Wrap(err, "fail to sign message")
	}
	hexTx := []byte(hex.EncodeToString(rawBz))
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
