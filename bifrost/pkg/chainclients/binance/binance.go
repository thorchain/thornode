package binance

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/binance-chain/go-sdk/common/types"
	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	ttypes "github.com/binance-chain/go-sdk/types"
	"github.com/binance-chain/go-sdk/types/msg"
	btx "github.com/binance-chain/go-sdk/types/tx"
	pkerrors "github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tssp "gitlab.com/thorchain/tss/go-tss/tss"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	pubkeymanager "gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
)

// Binance is a structure to sign and broadcast tx to binance chain used by signer mostly
type Binance struct {
	logger             zerolog.Logger
	RPCHost            string
	cfg                config.ChainConfiguration
	chainID            string
	isTestNet          bool
	client             *http.Client
	accts              *BinanceMetaDataStore
	currentBlockHeight int64
	tssKeyManager      keys.KeyManager
	localKeyManager    *keyManager
	thorchainBridge    *thorclient.ThorchainBridge
	storage            *BinanceBlockScannerStorage
	blockScanner       *BinanceBlockScanner
}

// NewBinance create new instance of binance client
func NewBinance(thorKeys *thorclient.Keys, cfg config.ChainConfiguration, server *tssp.TssServer, thorchainBridge *thorclient.ThorchainBridge) (*Binance, error) {
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("rpc host is empty")
	}
	rpcHost := cfg.RPCHost

	tssKm, err := tss.NewKeySign(server)
	if err != nil {
		return nil, fmt.Errorf("fail to create tss signer: %w", err)
	}

	priv, err := thorKeys.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("fail to get private key: %w", err)
	}

	pk, err := common.NewPubKeyFromCrypto(priv.PubKey())
	if err != nil {
		return nil, fmt.Errorf("fail to get pub key: %w", err)
	}
	if thorchainBridge == nil {
		return nil, errors.New("thorchain bridge is nil")
	}
	localKm := &keyManager{
		privKey: priv,
		addr:    ctypes.AccAddress(priv.PubKey().Address()),
		pubkey:  pk,
	}

	if !strings.HasPrefix(rpcHost, "http") {
		rpcHost = fmt.Sprintf("http://%s", rpcHost)
	}

	return &Binance{
		logger:          log.With().Str("module", "binance").Logger(),
		RPCHost:         rpcHost,
		cfg:             cfg,
		accts:           NewBinanceMetaDataStore(),
		client:          &http.Client{},
		tssKeyManager:   tssKm,
		localKeyManager: localKm,
		thorchainBridge: thorchainBridge,
	}, nil
}

func (b *Binance) initBlockScanner(pubkeyMgr pubkeymanager.PubKeyValidator, m *metrics.Metrics) error {
	b.checkIsTestNet()

	var err error
	b.storage, err = NewBinanceBlockScannerStorage(b.cfg.BlockScanner.DBPath)
	if err != nil {
		return pkerrors.Wrap(err, "fail to create scan storage")
	}
	startBlockHeight := int64(0)
	if !b.cfg.BlockScanner.EnforceBlockHeight {
		startBlockHeight, err = b.thorchainBridge.GetLastObservedInHeight(common.BNBChain)
		if err != nil {
			return pkerrors.Wrap(err, "fail to get start block height from thorchain")
		}
		if startBlockHeight == 0 {
			startBlockHeight, err = b.GetHeight()
			if err != nil {
				return pkerrors.Wrap(err, "fail to get binance height")
			}
			b.logger.Info().Int64("height", startBlockHeight).Msg("Current block height is indeterminate; using current height from Binance.")
		}
	} else {
		startBlockHeight = b.cfg.BlockScanner.StartBlockHeight
	}
	b.blockScanner, err = NewBinanceBlockScanner(b.cfg.BlockScanner, startBlockHeight, b.storage, b.isTestNet, pubkeyMgr, m)
	if err != nil {
		return pkerrors.Wrap(err, "fail to create block scanner")
	}
	return nil
}

func (b *Binance) Start(globalTxsQueue chan stypes.TxIn, pubkeyMgr pubkeymanager.PubKeyValidator, m *metrics.Metrics) error {
	err := b.initBlockScanner(pubkeyMgr, m)
	if err != nil {
		b.logger.Error().Err(err).Msg("fail to init block scanner")
		return err
	}
	b.blockScanner.Start(globalTxsQueue)
	return nil
}

func (b *Binance) Stop() error {
	return b.blockScanner.Stop()
}

// IsTestNet determinate whether we are running on test net by checking the status
func (b *Binance) checkIsTestNet() {
	// Cached data after first call
	if b.isTestNet {
		return
	}

	u, err := url.Parse(b.RPCHost)
	if err != nil {
		log.Fatal().Msgf("Unable to parse rpc host: %s\n", b.RPCHost)
	}

	u.Path = "/status"

	resp, err := b.client.Get(u.String())
	if err != nil {
		log.Fatal().Msgf("%v\n", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
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
	if err := json.Unmarshal(data, &status); err != nil {
		log.Fatal().Err(err).Msg("fail to unmarshal body")
	}

	b.chainID = status.Result.NodeInfo.Network
	b.isTestNet = b.chainID == "Binance-Chain-Nile"

	if b.isTestNet {
		types.Network = types.TestNetwork
	} else {
		types.Network = types.ProdNetwork
	}
}

func (b *Binance) GetChain() common.Chain {
	return common.BNBChain
}

func (b *Binance) GetHeight() (int64, error) {
	u, err := url.Parse(b.RPCHost)
	if err != nil {
		return 0, fmt.Errorf("unable to parse dex host: %w", err)
	}
	u.Path = "abci_info"
	resp, err := b.client.Get(u.String())
	if err != nil {
		return 0, fmt.Errorf("fail to get request(%s): %w", u.String(), err) // errors.Wrap(err, "Get request failed")
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error().Err(err).Msg("fail to close resp body")
		}
	}()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("fail to read resp body: %w", err)
	}

	type ABCIinfo struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      string `json:"id"`
		Result  struct {
			Response struct {
				BlockHeight string `json:"last_block_height"`
			} `json:"response"`
		} `json:"result"`
	}

	var abci ABCIinfo
	if err := json.Unmarshal(data, &abci); err != nil {
		return 0, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return strconv.ParseInt(abci.Result.Response.BlockHeight, 10, 64)
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
	if err != nil {
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
	addr, err := poolPubKey.GetAddress(common.BNBChain)
	if err != nil {
		b.logger.Error().Err(err).Str("pool_pub_key", poolPubKey.String()).Msg("fail to get pool address")
		return ""
	}
	return addr.String()
}

func (b *Binance) GetGasFee(count uint64) common.Gas {
	return common.GetBNBGasFee(count)
}

func (b *Binance) ValidateMetadata(inter interface{}) bool {
	meta := inter.(BinanceMetadata)
	acct := b.accts.GetByAccount(meta.AccountNumber)
	return acct.AccountNumber == meta.AccountNumber && acct.SeqNumber == meta.SeqNumber
}

// SignTx sign the the given TxArrayItem
func (b *Binance) SignTx(tx stypes.TxOutItem, height int64) ([]byte, error) {
	var payload []msg.Transfer

	toAddr, err := types.AccAddressFromBech32(tx.ToAddress.String())
	if err != nil {
		return nil, fmt.Errorf("fail to parse account address(%s) :%w", tx.ToAddress.String(), err)
	}

	var coins types.Coins
	for _, coin := range tx.Coins {
		coins = append(coins, types.Coin{
			Denom:  coin.Asset.Symbol.String(),
			Amount: int64(coin.Amount.Uint64()),
		})
	}

	payload = append(payload, msg.Transfer{
		ToAddr: toAddr,
		Coins:  coins,
	})

	if len(payload) == 0 {
		b.logger.Error().Msg("payload is empty , this should not happen")
		return nil, nil
	}
	fromAddr := b.GetAddress(tx.VaultPubKey)
	sendMsg := b.parseTx(fromAddr, payload)
	if err := sendMsg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("invalid send msg: %w", err)
	}

	currentHeight, err := b.GetHeight()
	if err != nil {
		b.logger.Error().Err(err).Msg("fail to get current binance block height")
		return nil, err
	}
	if currentHeight > b.currentBlockHeight {
		acc, err := b.GetAccount(fromAddr)
		if err != nil {
			return nil, fmt.Errorf("fail to get account info: %w", err)
		}
		atomic.StoreInt64(&b.currentBlockHeight, currentHeight)
		b.accts.Set(tx.VaultPubKey, BinanceMetadata{
			AccountNumber: acc.AccountNumber,
			SeqNumber:     acc.Sequence,
		})
	}
	meta := b.accts.Get(tx.VaultPubKey)
	b.logger.Info().Int64("account_number", meta.AccountNumber).Int64("sequence_number", meta.SeqNumber).Msg("account info")
	signMsg := btx.StdSignMsg{
		ChainID:       b.chainID,
		Memo:          tx.Memo,
		Msgs:          []msg.Msg{sendMsg},
		Source:        btx.Source,
		Sequence:      meta.SeqNumber,
		AccountNumber: meta.AccountNumber,
	}
	rawBz, err := b.signWithRetry(signMsg, fromAddr, tx.VaultPubKey, height, tx)
	if err != nil {
		return nil, fmt.Errorf("fail to sign message: %w", err)
	}

	if len(rawBz) == 0 {
		// the transaction was already signed
		return nil, nil
	}

	// increment sequence number
	b.accts.SeqInc(tx.VaultPubKey)

	hexTx := []byte(hex.EncodeToString(rawBz))
	return hexTx, nil
}

func (b *Binance) sign(signMsg btx.StdSignMsg, poolPubKey common.PubKey, signerPubKeys common.PubKeys) ([]byte, error) {
	if b.localKeyManager.Pubkey().Equals(poolPubKey) {
		return b.localKeyManager.Sign(signMsg)
	}
	k := b.tssKeyManager.(tss.ThorchainKeyManager)
	return k.SignWithPool(signMsg, poolPubKey, signerPubKeys)
}

// signWithRetry is design to sign a given message until it success or the same message had been send out by other signer
func (b *Binance) signWithRetry(signMsg btx.StdSignMsg, from string, poolPubKey common.PubKey, height int64, txOutItem stypes.TxOutItem) ([]byte, error) {
	for {
		keySignParty, err := b.thorchainBridge.GetKeysignParty(poolPubKey)
		if err != nil {
			b.logger.Error().Err(err).Msg("fail to get keysign party")
			continue
		}

		// We get the keysign object from thorchain again to ensure it hasn't
		// been signed already, and we can skip. This helps us not get stuck on
		// a task that we'll never sign, because 2/3rds already has and will
		// never be available to sign again.
		txOut, err := b.thorchainBridge.GetKeysign(height, poolPubKey.String())
		if err != nil {
			b.logger.Error().Err(err).Msg("fail to get keysign items")
			continue
		}
		for _, out := range txOut.Chains {
			for _, tx := range out.TxArray {
				item := tx.TxOutItem()
				if txOutItem.Equals(item) && !tx.OutHash.IsEmpty() {
					// already been signed, we can skip it
					b.logger.Info().Str("tx_id", tx.OutHash.String()).Msgf("already signed. skippping...")
					return nil, nil
				}
			}
		}

		rawBytes, err := b.sign(signMsg, poolPubKey, keySignParty)
		if err == nil && rawBytes != nil {
			return rawBytes, nil
		}
		var keysignError tss.KeysignError
		if errors.As(err, &keysignError) {
			if len(keysignError.Blame.BlameNodes) == 0 {
				// TSS doesn't know which node to blame
				continue
			}

			// key sign error forward the keysign blame to thorchain
			txID, err := b.thorchainBridge.PostKeysignFailure(keysignError.Blame, height, txOutItem.Memo, txOutItem.Coins)
			if err != nil {
				b.logger.Error().Err(err).Msg("fail to post keysign failure to thorchain")
			} else {
				b.logger.Info().Str("tx_id", txID.String()).Msgf("post keysign failure to thorchain")
			}
			continue
		}
		b.logger.Error().Err(err).Msgf("fail to sign msg with memo: %s", signMsg.Memo)
		// should THORNode give up? let's check the seq no on binance chain
		// keep in mind, when THORNode don't run our own binance full node, THORNode might get rate limited by binance

		acc, err := b.GetAccount(from)
		if err != nil {
			b.logger.Error().Err(err).Msg("fail to get account info from binance chain")
			continue
		}
		if acc.Sequence > signMsg.Sequence {
			b.logger.Debug().Msgf("msg with memo: %s , seqNo: %d had been processed", signMsg.Memo, signMsg.Sequence)
			return nil, nil
		}
	}
}

func (b *Binance) GetAccount(addr string) (common.Account, error) {
	address, err := types.AccAddressFromBech32(addr)
	if err != nil {
		b.logger.Error().Err(err).Msgf("fail to get parse address: %s", addr)
		return common.Account{}, err
	}
	u, err := url.Parse(b.RPCHost)
	if err != nil {
		log.Fatal().Msgf("Error parsing rpc (%s): %s", b.RPCHost, err)
		return common.Account{}, err
	}
	u.Path = "/abci_query"
	v := u.Query()
	v.Set("path", fmt.Sprintf("\"/account/%s\"", address.String()))
	u.RawQuery = v.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return common.Account{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
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
		return common.Account{}, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return common.Account{}, err
	}

	data, err := base64.StdEncoding.DecodeString(result.Result.Response.Value)
	if err != nil {
		return common.Account{}, err
	}

	cdc := ttypes.NewCodec()
	var acc types.AppAccount
	err = cdc.UnmarshalBinaryBare(data, &acc)
	if err != nil {
		return common.Account{}, err
	}
	account := common.NewAccount(acc.BaseAccount.Sequence, acc.BaseAccount.AccountNumber, common.GetCoins(acc.BaseAccount.Coins))
	return account, nil
}

// broadcastTx is to broadcast the tx to binance chain
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
		return fmt.Errorf("fail to broadcast tx to binance chain: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		result, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("fail to read response body: %w", err)
		}
		log.Info().Msg(string(result))
		return fmt.Errorf("fail to broadcast tx to binance:(%s)", b.RPCHost)
	}
	err = resp.Body.Close()
	if err != nil {
		log.Error().Err(err).Msg("we fail to close response body")
		return errors.New("fail to close response body")
	}
	return nil
}
