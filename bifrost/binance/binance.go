package binance

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/binance-chain/go-sdk/client/basic"
	"github.com/binance-chain/go-sdk/client/query"
	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/types/tx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/crypto"

	"gitlab.com/thorchain/bepswap/thornode/bifrost/config"
	stypes "gitlab.com/thorchain/bepswap/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/bepswap/thornode/bifrost/tss"
	"gitlab.com/thorchain/bepswap/thornode/common"
)

type Binance struct {
	logger      zerolog.Logger
	cfg         config.BinanceConfiguration
	basicClient basic.BasicClient
	queryClient query.QueryClient
	keyManager  keys.KeyManager
	chainId     string
	useTSS      bool
}

// NewBinance create new instance of binance client
func NewBinance(cfg config.BinanceConfiguration, useTSS bool, keySignCfg config.TSSConfiguration) (*Binance, error) {
	if !useTSS && len(cfg.PrivateKey) == 0 {
		return nil, errors.New("no private key")
	}
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("dex host is empty, set env DEX_HOST")
	}
	var km keys.KeyManager
	var err error
	if useTSS {
		km, err = tss.NewKeySign(keySignCfg)
		if nil != err {
			return nil, errors.Wrap(err, "fail to create tss signer")
		}
	} else {
		km, err = keys.NewPrivateKeyManager(cfg.PrivateKey)
		if err != nil {
			return nil, errors.Wrap(err, "fail to create private key manager")
		}
	}

	host := cfg.RPCHost
	// drop http/https prefix
	if strings.HasPrefix(cfg.RPCHost, "http://") {
		host = cfg.RPCHost[7:]
	}
	if strings.HasPrefix(cfg.RPCHost, "https://") {
		host = cfg.RPCHost[8:]
	}

	basicClient := basic.NewClient(host)
	queryClient := query.NewClient(basicClient)
	return &Binance{
		logger:      log.With().Str("module", "binance").Logger(),
		cfg:         cfg,
		basicClient: basicClient,
		queryClient: queryClient,
		keyManager:  km,
		chainId:     "Binance-Chain-Nile", // TODO: this should be configurable
		useTSS:      useTSS,
	}, nil
}

func IsTestNet(dexHost string) bool {
	client := &http.Client{}

	u, err := url.Parse(dexHost)
	if err != nil {
		log.Fatal().Msgf("Unable to parse dex host: %s\n", dexHost)
	}

	uri := url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   "/status",
	}

	resp, err := client.Get(uri.String())
	if err != nil {
		log.Fatal().Msgf("%v\n", err)
	}

	defer func() {
		if err := resp.Body.Close(); nil != err {
			log.Error().Err(err).Msg("fail to close resp body")
		}
	}()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err)
	}

	type Status struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      string `json:"id"`
		Result  struct {
			NodeInfo struct {
				Network string `json:"network"`
			} `json:"node_info"`
		} `json:"result"`
	}

	var status Status
	if err := json.Unmarshal(data, &status); nil != err {
		log.Error().Err(err)
	}

	return status.Result.NodeInfo.Network == "Binance-Chain-Nile"
}

func (b *Binance) input(addr types.AccAddress, coins types.Coins) msg.Input {
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
	input := b.input(from, fromCoins)
	output := make([]msg.Output, 0, len(transfers))
	for _, t := range transfers {
		t.Coins = t.Coins.Sort()
		output = append(output, b.output(t.ToAddr, t.Coins))
	}
	return b.msgToSend([]msg.Input{input}, output)
}

func (b *Binance) parseTx(fromAddr string, transfers []msg.Transfer) msg.SendMsg {
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

// GetAddress return current signer address, it will be bech32 encoded address
func (b *Binance) GetAddress(poolPubKey common.PubKey) string {
	if !b.useTSS {
		return b.keyManager.GetAddr().String()
	}
	addr, err := poolPubKey.GetAddress(common.BNBChain)
	if nil != err {
		b.logger.Error().Err(err).Str("pool_pub_key", poolPubKey.String()).Msg("fail to get pool address")
		return ""
	}
	return addr.String()
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
	signerAddr := b.GetAddress(tai.PoolAddress)
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

	payload = append(payload, msg.Transfer{
		ToAddr: toAddr,
		Coins: types.Coins{
			types.Coin{
				Denom:  tai.Coin.Asset.Symbol.String(),
				Amount: int64(tai.Coin.Amount.Uint64()),
			},
		},
	})

	if len(payload) == 0 {
		b.logger.Error().Msg("payload is empty , this should not happen")
		return nil, nil, nil
	}
	fromAddr := b.GetAddress(tai.PoolAddress)
	sendMsg := b.parseTx(fromAddr, payload)
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
	rawBz, err := b.signWithRetry(signMsg, fromAddr)
	if nil != err {
		return nil, nil, errors.Wrap(err, "fail to sign message")
	}

	if len(rawBz) == 0 {
		return nil, nil, nil
	}
	hexTx := []byte(hex.EncodeToString(rawBz))
	return hexTx, param, nil
}

// signWithRetry is design to sign a given message until it success or the same message had been send out by other signer
func (b *Binance) signWithRetry(signMsg tx.StdSignMsg, from string) ([]byte, error) {
	for {
		rawBytes, err := b.keyManager.Sign(signMsg)
		if nil == err {
			return rawBytes, nil
		}
		b.logger.Error().Err(err).Msgf("fail to sign msg with memo: %s", signMsg.Memo)
		// should we give up? let's check the seq no on binance chain
		// keep in mind, when we don't run our own binance full node, we might get rate limited by binance
		acc, err := b.queryClient.GetAccount(from)
		if nil != err {
			b.logger.Error().Err(err).Msg("fail to get account info from binance chain")
			continue
		}
		if acc.Sequence > signMsg.Sequence {
			b.logger.Debug().Msgf("msg with memo: %s , seqNo: %d had been processed", signMsg.Memo, signMsg.Sequence)
			return nil, nil
		}
	}
}

func (b *Binance) BroadcastTx(hexTx []byte, param map[string]string) (*tx.TxCommitResult, error) {
	commits, err := b.basicClient.PostTx(hexTx, param)
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

// GetPubKey return the pub key
func (b *Binance) GetPubKey() crypto.PubKey {
	return b.keyManager.GetPrivKey().PubKey()
}
