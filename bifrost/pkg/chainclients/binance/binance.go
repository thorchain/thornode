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
	"sync"
	"sync/atomic"
	"time"

	"github.com/binance-chain/go-sdk/client/rpc"
	"github.com/binance-chain/go-sdk/common/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/binance-chain/go-sdk/keys"
	ttypes "github.com/binance-chain/go-sdk/types"
	"github.com/binance-chain/go-sdk/types/msg"
	bmsg "github.com/binance-chain/go-sdk/types/msg"
	btx "github.com/binance-chain/go-sdk/types/tx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tssp "gitlab.com/thorchain/tss/go-tss/tss"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
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
	pubkeyMgr          pubkeymanager.PubKeyValidator
	client             rpc.Client
	accountNumber      int64
	wg                 *sync.WaitGroup
	seqNumber          int64
	currentBlockHeight int64
	signLock           *sync.Mutex
	tssKeyManager      keys.KeyManager
	localKeyManager    *keyManager
	thorchainBridge    *thorclient.ThorchainBridge
	stopChan           chan struct{}
}

type BinanceMetadata struct {
	AccountNumber int64
	SeqNumber     int64
}

// NewBinance create new instance of binance client
func NewBinance(thorKeys *thorclient.Keys, cfg config.ChainConfiguration, server *tssp.TssServer, thorchainBridge *thorclient.ThorchainBridge) (*Binance, error) {
	if len(cfg.RPCHost) == 0 {
		return nil, errors.New("rpc host is empty")
	}
	rpcHost := cfg.RPCHost

	network := setNetwork(cfg)

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
		addr:    types.AccAddress(priv.PubKey().Address()),
		pubkey:  pk,
	}

	if !strings.HasPrefix(rpcHost, "http") {
		rpcHost = fmt.Sprintf("http://%s", rpcHost)
	}

	return &Binance{
		logger:          log.With().Str("module", "binance").Logger(),
		RPCHost:         rpcHost,
		cfg:             cfg,
		client:          rpc.NewRPCClient(rpcHost, network),
		signLock:        &sync.Mutex{},
		tssKeyManager:   tssKm,
		localKeyManager: localKm,
		wg:              &sync.WaitGroup{},
		stopChan:        make(chan struct{}),
		thorchainBridge: thorchainBridge,
	}, nil
}

func setNetwork(cfg config.ChainConfiguration) types.ChainNetwork {
	var network types.ChainNetwork
	if cfg.ChainNetwork == strings.ToLower("mainnet") {
		network = types.ProdNetwork
	}

	if cfg.ChainNetwork == strings.ToLower("testnet") || cfg.ChainNetwork == "" {
		network = types.TestNetwork
	}
	return network
}

func (b *Binance) initScanBlockHeight() (err error) {
	if !b.cfg.BlockScanner.EnforceBlockHeight {
		b.currentBlockHeight, err = b.thorchainBridge.GetLastObservedInHeight(common.BNBChain)
		if err != nil {
			return errors.Wrap(err, "fail to get start block height from thorchain")
		}
		if b.currentBlockHeight == 0 {
			b.currentBlockHeight, err = b.GetHeight()
			if err != nil {
				return errors.Wrap(err, "fail to get binance height")
			}
			b.logger.Info().Int64("height", b.currentBlockHeight).Msg("Current block height is indeterminate; using current height from Binance.")
		}
	} else {
		b.currentBlockHeight = b.cfg.BlockScanner.StartBlockHeight
	}
	return nil
}

// Start starts scanning blocks on binance chain
func (b *Binance) Start(globalTxsQueue chan stypes.TxIn, pubkeyMgr pubkeymanager.PubKeyValidator, m *metrics.Metrics) error {

	b.pubkeyMgr = pubkeyMgr

	err := b.initScanBlockHeight()
	if err != nil {
		b.logger.Error().Err(err).Msg("fail to init block scanner height")
		return err
	}
	b.wg.Add(1)
	go b.scanBlocks(globalTxsQueue)
	return nil
}

func (b *Binance) scanBlocks(globalTxsQueue chan stypes.TxIn) {
	for {
		select {
		case <-b.stopChan:
			return
		default:
			block, err := b.GetBlock(&b.currentBlockHeight)
			if err != nil || block.Block == nil {
				b.logger.Debug().Int64("block height", b.currentBlockHeight).Msg("backing off before getting next block")
				time.Sleep(b.cfg.BlockScanner.BlockHeightDiscoverBackoff)
				continue
			}
			globalTxsQueue <- b.processBlock(block)
			b.currentBlockHeight++
		}
	}
}

// processBlock extract's block information of interest into out generic/simplified block struct to processing by the txBlockScanner module.
func (b *Binance) processBlock(block *ctypes.ResultBlock) (txIn stypes.TxIn) {
	for _, tx := range block.Block.Data.Txs {
		var t btx.StdTx
		if err := btx.Cdc.UnmarshalBinaryLengthPrefixed(tx, &t); err != nil {
			b.logger.Err(err).Msg("UnmarshalBinaryLengthPrefixed")
		}

		hash := fmt.Sprintf("%X", tx.Hash())
		txItems, err := b.processStdTx(hash, t)
		if err != nil {
			b.logger.Err(err).Msg("failed to processStdTx")
			continue
		}

		// if valid txItems returned
		if len(txItems) > 0 {
			txIn.TxArray = append(txIn.TxArray, txItems...)
		}
	}
	txIn.BlockHeight = strconv.FormatInt(block.Block.Header.Height, 10)
	txIn.Count = strconv.Itoa(len(txIn.TxArray))
	txIn.Chain = common.BNBChain
	return txIn
}

// processStdTx extract's tx information of interest into our generic TxItem struct
func (b *Binance) processStdTx(hash string, stdTx btx.StdTx) (txItems []stypes.TxInItem, err error) {
	// TODO: it is possible to have multiple `SendMsg` in a single stdTx, which
	// THORNode are currently not accounting for. It is also possible to have
	// multiple inputs/outputs within a single stdTx, which THORNode are not yet
	// accounting for.
	for _, msg := range stdTx.Msgs {
		switch sendMsg := msg.(type) {
		case bmsg.SendMsg:
			txInItem := stypes.TxInItem{
				Tx: hash,
			}
			txInItem.Memo = stdTx.Memo

			// THORNode take the first Input as sender, first Output as receiver
			// so if THORNode send to multiple different receiver within one tx, this won't be able to process it.
			sender := sendMsg.Inputs[0]
			receiver := sendMsg.Outputs[0]
			txInItem.Sender = sender.Address.String()
			txInItem.To = receiver.Address.String()

			txInItem.Coins, err = b.getCoinsForTxIn(sendMsg.Outputs)
			if err != nil {
				return nil, errors.Wrap(err, "fail to convert coins")
			}

			// TODO: We should not assume what the gas fees are going to be in
			// the future (although they are largely static for binance). We
			// should modulus the binance block height and get the latest fee
			// prices every 1,000 or so blocks. This would ensure that all
			// observers will always report the same gas prices as they update
			// their price fees at the same time.

			// Calculate gas for this tx
			if len(txInItem.Coins) > 1 {
				// Multisend gas fees
				txInItem.Gas = common.GetBNBGasFeeMulti(uint64(len(txInItem.Coins)))
			} else {
				// Single transaction gas fees
				txInItem.Gas = common.BNBGasFeeSingleton
			}

			if ok := b.MatchedAddress(txInItem); !ok {
				continue
			}

			// NOTE: the following could result in the same tx being added
			// twice, which is expected. We want to make sure we generate both
			// a inbound and outbound txn, if we both apply.

			// check if the from address is a valid pool
			if ok, cpi := b.pubkeyMgr.IsValidPoolAddress(txInItem.Sender, common.BNBChain); ok {
				txInItem.ObservedPoolAddress = cpi.PubKey.String()
				txItems = append(txItems, txInItem)
			}
			// check if the to address is a valid pool address
			if ok, cpi := b.pubkeyMgr.IsValidPoolAddress(txInItem.To, common.BNBChain); ok {
				txInItem.ObservedPoolAddress = cpi.PubKey.String()
				txItems = append(txItems, txInItem)
			} else {
				// Apparently we don't recognize where we are sending funds to.
				// Lets check if we should because its an internal transaction
				// moving funds between vaults (for example). If it is, lets
				// manually trigger an update of pubkeys, then check again...
				switch strings.ToLower(txInItem.Memo) {
				case "migrate", "yggdrasil-", "yggdrasil+":
					b.pubkeyMgr.FetchPubKeys()
					if ok, cpi := b.pubkeyMgr.IsValidPoolAddress(txInItem.To, common.BNBChain); ok {
						txInItem.ObservedPoolAddress = cpi.PubKey.String()
						txItems = append(txItems, txInItem)
					}
				}
			}
		default:
			continue
		}
	}
	return txItems, nil
}

// MatchedAddress checks addresses match our pool addresses
func (b *Binance) MatchedAddress(txInItem stypes.TxInItem) bool {
	// Check if we are migrating our funds...
	if ok := b.isMigration(txInItem.Sender, txInItem.Memo); ok {
		b.logger.Debug().Str("memo", txInItem.Memo).Msg("migrate")
		return true
	}

	// Check if our pool is registering a new yggdrasil pool. Ie
	// sending the staked assets to the user
	if ok := b.isRegisterYggdrasil(txInItem.Sender, txInItem.Memo); ok {
		b.logger.Debug().Str("memo", txInItem.Memo).Msg("yggdrasil+")
		return true
	}

	// Check if out pool is de registering a yggdrasil pool. Ie sending
	// the bond back to the user
	if ok := b.isDeregisterYggdrasil(txInItem.Sender, txInItem.Memo); ok {
		b.logger.Debug().Str("memo", txInItem.Memo).Msg("yggdrasil-")
		return true
	}

	// Check if THORNode are sending from a yggdrasil address
	if ok := b.isYggdrasil(txInItem.Sender); ok {
		b.logger.Debug().Str("assets sent from yggdrasil pool", txInItem.Memo).Msg("fill order")
		return true
	}

	// Check if THORNode are sending to a yggdrasil address
	if ok := b.isYggdrasil(txInItem.To); ok {
		b.logger.Debug().Str("assets to yggdrasil pool", txInItem.Memo).Msg("refill")
		return true
	}

	// outbound message from pool, when it is outbound, it does not matter how much coins THORNode send to customer for now
	if ok := b.isOutboundMsg(txInItem.Sender, txInItem.Memo); ok {
		b.logger.Debug().Str("memo", txInItem.Memo).Msg("outbound")
		return true
	}

	return false
}

// Check if memo is for registering an Asgard vault
func (b *Binance) isMigration(addr, memo string) bool {
	return b.isAddrWithMemo(addr, memo, "migrate")
}

// Check if memo is for registering a Yggdrasil vault
func (b *Binance) isRegisterYggdrasil(addr, memo string) bool {
	return b.isAddrWithMemo(addr, memo, "yggdrasil+")
}

// Check if memo is for de registering a Yggdrasil vault
func (b *Binance) isDeregisterYggdrasil(addr, memo string) bool {
	return b.isAddrWithMemo(addr, memo, "yggdrasil-")
}

// Check if THORNode have an outbound yggdrasil transaction
func (b *Binance) isYggdrasil(addr string) bool {
	ok, _ := b.pubkeyMgr.IsValidPoolAddress(addr, common.BNBChain)
	return ok
}

func (b *Binance) isOutboundMsg(addr, memo string) bool {
	return b.isAddrWithMemo(addr, memo, "outbound")
}

func (b *Binance) isAddrWithMemo(addr, memo, targetMemo string) bool {
	match, _ := b.pubkeyMgr.IsValidPoolAddress(addr, common.BNBChain)
	if !match {
		return false
	}
	lowerMemo := strings.ToLower(memo)
	if strings.HasPrefix(lowerMemo, targetMemo) {
		return true
	}
	return false
}

// getCoinsForTxIn extract's the coins/amount into our generic Coins struct
func (b *Binance) getCoinsForTxIn(outputs []bmsg.Output) (common.Coins, error) {
	cc := common.Coins{}
	for _, output := range outputs {
		for _, c := range output.Coins {
			asset, err := common.NewAsset(fmt.Sprintf("BNB.%s", c.Denom))
			if err != nil {
				return nil, errors.Wrapf(err, "fail to create asset, %s is not valid", c.Denom)
			}
			amt := sdk.NewUint(uint64(c.Amount))
			cc = append(cc, common.NewCoin(asset, amt))
		}
	}
	return cc, nil
}

// GetBlock gets the block for a height
func (b *Binance) GetBlock(blockHeight *int64) (*ctypes.ResultBlock, error) {
	return b.client.Block(blockHeight)
}

// Stop stops scanning incoming block on binance chain
func (b *Binance) Stop() error {
	b.logger.Debug().Msg("receive stop request")
	defer b.logger.Debug().Msg("binance block scanner stopped")
	close(b.stopChan)
	b.wg.Wait()
	return nil
}

// GetChain returns chain object
func (b *Binance) GetChain() common.Chain {
	return common.BNBChain
}

// GetHeight return current block height in binance chain
func (b *Binance) GetHeight() (int64, error) {
	block, err := b.GetBlock(nil)
	if err != nil {
		return 0, errors.Wrap(err, "unable to retrieve block from binance")
	}
	return block.Block.Header.Height, nil
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
	return meta.AccountNumber == b.accountNumber && meta.SeqNumber == b.seqNumber
}

// SignTx sign the the given TxArrayItem
func (b *Binance) SignTx(tx stypes.TxOutItem, height int64) ([]byte, error) {
	b.signLock.Lock()
	defer b.signLock.Unlock()
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
		atomic.StoreInt64(&b.accountNumber, acc.AccountNumber)
		atomic.StoreInt64(&b.seqNumber, acc.Sequence)
	}
	b.logger.Info().Int64("account_number", b.accountNumber).Int64("sequence_number", b.seqNumber).Msg("account info")
	signMsg := btx.StdSignMsg{
		ChainID:       b.chainID,
		Memo:          tx.Memo,
		Msgs:          []msg.Msg{sendMsg},
		Source:        btx.Source,
		Sequence:      b.seqNumber,
		AccountNumber: b.accountNumber,
	}
	rawBz, err := b.signWithRetry(signMsg, fromAddr, tx.VaultPubKey, height, tx.Memo, tx.Coins)
	if err != nil {
		return nil, fmt.Errorf("fail to sign message: %w", err)
	}

	if len(rawBz) == 0 {
		// this could happen, if the local party trying to sign a message , however the TSS keysign process didn't chose the local party to sign the message
		return nil, nil
	}

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
func (b *Binance) signWithRetry(signMsg btx.StdSignMsg, from string, poolPubKey common.PubKey, height int64, memo string, coins common.Coins) ([]byte, error) {
	for {
		keySignParty, err := b.thorchainBridge.GetKeysignParty(poolPubKey)
		if err != nil {
			b.logger.Error().Err(err).Msg("fail to get keysign party")
			continue
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
			txID, err := b.thorchainBridge.PostKeysignFailure(keysignError.Blame, height, memo, coins)
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
	atomic.AddInt64(&b.seqNumber, 1)
	return nil
}
