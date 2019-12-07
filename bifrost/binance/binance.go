package binance

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	ttypes "github.com/binance-chain/go-sdk/types"
	"github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/types/tx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/crypto"

	"gitlab.com/thorchain/thornode/bifrost/config"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
)

// Binance is a structure to sign and broadcast tx to binance chain used by signer mostly
type Binance struct {
	logger     zerolog.Logger
	cfg        config.BinanceConfiguration
	keyManager keys.KeyManager
	RPCHost    string
	chainID    string
	useTSS     bool
	isTestNet  bool
}

// NewBinance create new instance of binance client
func NewBinance(cfg config.BinanceConfiguration, useTSS bool, keySignCfg config.TSSConfiguration) (*Binance, error) {
	if !useTSS && len(cfg.PrivateKey) == 0 {
		return nil, errors.New("no private key")
	}
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("rpc host is empty")
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

	rpcHost := cfg.RPCHost
	if !strings.HasPrefix(rpcHost, "http") {
		rpcHost = fmt.Sprintf("http://%s", rpcHost)
	}

	chainID, isTestNet := IsTestNet(rpcHost)
	if isTestNet {
		types.Network = types.TestNetwork
	} else {
		types.Network = types.ProdNetwork
	}

	return &Binance{
		logger:     log.With().Str("module", "binance").Logger(),
		cfg:        cfg,
		keyManager: km,
		RPCHost:    rpcHost,
		chainID:    chainID,
		isTestNet:  isTestNet,
		useTSS:     useTSS,
	}, nil
}

// IsTestNet determinate whether we are running on test net by checking the status
func IsTestNet(rpcHost string) (string, bool) {
	client := &http.Client{}

	u, err := url.Parse(rpcHost)
	if err != nil {
		log.Fatal().Msgf("Unable to parse rpc host: %s\n", rpcHost)
	}

	u.Path = "/status"

	resp, err := client.Get(u.String())
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
		log.Fatal().Err(err).Msg("fail to read body")
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
		log.Fatal().Err(err).Msg("fail to unmarshal body")
	}

	isTestNet := status.Result.NodeInfo.Network == "Binance-Chain-Nile"
	return status.Result.NodeInfo.Network, isTestNet
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

// SignTx sign the the given TxArrayItem
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

	address, err := types.AccAddressFromBech32(fromAddr)
	if err != nil {
		b.logger.Error().Err(err).Msgf("fail to get parse address: %s", fromAddr)
		return nil, nil, err
	}

	acc, err := b.GetAccount(address)
	if err != nil {
		return nil, nil, errors.Wrap(err, "fail to get account info")
	}

	signMsg := tx.StdSignMsg{
		ChainID:       b.chainID,
		Memo:          tai.Memo,
		Msgs:          []msg.Msg{sendMsg},
		Source:        tx.Source,
		Sequence:      seqNo, // acc.Sequence,
		AccountNumber: acc.AccountNumber,
	}
	param := map[string]string{
		"sync": "true",
	}
	rawBz, err := b.signWithRetry(signMsg, fromAddr, tai.PoolAddress)
	if nil != err {
		return nil, nil, errors.Wrap(err, "fail to sign message")
	}

	if len(rawBz) == 0 {
		return nil, nil, nil
	}
	hexTx := []byte(hex.EncodeToString(rawBz))
	return hexTx, param, nil
}

func (b *Binance) sign(signMsg tx.StdSignMsg, poolPubKey common.PubKey) ([]byte, error) {
	if b.useTSS {
		k := b.keyManager.(tss.ThorchainKeyManager)
		return k.SignWithPool(signMsg, poolPubKey)
	}
	return b.keyManager.Sign(signMsg)
}

// signWithRetry is design to sign a given message until it success or the same message had been send out by other signer
func (b *Binance) signWithRetry(signMsg tx.StdSignMsg, from string, poolPubKey common.PubKey) ([]byte, error) {
	for {
		rawBytes, err := b.sign(signMsg, poolPubKey)
		if nil == err {
			return rawBytes, nil
		}
		b.logger.Error().Err(err).Msgf("fail to sign msg with memo: %s", signMsg.Memo)
		// should THORNode give up? let's check the seq no on binance chain
		// keep in mind, when THORNode don't run our own binance full node, THORNode might get rate limited by binance
		address, err := types.AccAddressFromBech32(from)
		if err != nil {
			b.logger.Error().Err(err).Msgf("fail to get parse address: %s", from)
			return nil, err
		}

		acc, err := b.GetAccount(address)
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

func (b *Binance) GetAccount(addr types.AccAddress) (types.BaseAccount, error) {
	u, err := url.Parse(b.RPCHost)
	if err != nil {
		log.Fatal().Msgf("Error parsing rpc (%s): %s", b.RPCHost, err)
		return types.BaseAccount{}, err
	}
	u.Path = "/abci_query"
	v := u.Query()
	v.Set("path", fmt.Sprintf("\"/account/%s\"", addr.String()))
	u.RawQuery = v.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return types.BaseAccount{}, err
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			b.logger.Error().Err(err).Msg("fail to close response body")
		}
	}()

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
		return types.BaseAccount{}, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return types.BaseAccount{}, err
	}

	data, err := base64.StdEncoding.DecodeString(result.Result.Response.Value)
	if err != nil {
		return types.BaseAccount{}, err
	}

	cdc := ttypes.NewCodec()
	var acc types.AppAccount
	err = cdc.UnmarshalBinaryBare(data, &acc)

	return acc.BaseAccount, err
}

// BroadcastTx is to broadcast the tx to binance chain
func (b *Binance) BroadcastTx(hexTx []byte) error {
	u, err := url.Parse(b.RPCHost)
	if err != nil {
		log.Error().Msgf("Error parsing rpc (%s): %s", b.RPCHost, err)
		return err
	}
	u.Path = "broadcast_tx_commit"
	values := u.Query()
	values.Set("tx", "0x"+string(hexTx))
	u.RawQuery = values.Encode()
	resp, err := http.Post(u.String(), "", nil)
	if err != nil {
		return errors.Wrap(err, "fail to broadcast tx to ")
	}
	defer func() {
		if err := resp.Body.Close(); nil != err {
			log.Error().Err(err).Msg("we fail to close response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		result, err := ioutil.ReadAll(resp.Body)
		if nil != err {
			return fmt.Errorf("fail to read response body: %w", err)
		}
		log.Info().Msg(string(result))
		return fmt.Errorf("fail to broadcast tx to binance:(%s)", b.RPCHost)
	}
	return nil
}

// GetPubKey return the pub key
func (b *Binance) GetPubKey() crypto.PubKey {
	return b.keyManager.GetPrivKey().PubKey()
}
