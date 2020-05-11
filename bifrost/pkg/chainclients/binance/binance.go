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

	"github.com/binance-chain/go-sdk/common/types"
	ctypes "github.com/binance-chain/go-sdk/common/types"
	"github.com/binance-chain/go-sdk/keys"
	ttypes "github.com/binance-chain/go-sdk/types"
	"github.com/binance-chain/go-sdk/types/msg"
	btx "github.com/binance-chain/go-sdk/types/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tssp "gitlab.com/thorchain/tss/go-tss/tss"

	"gitlab.com/thorchain/thornode/bifrost/blockscanner"
	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/metrics"
	"gitlab.com/thorchain/thornode/bifrost/thorclient"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/bifrost/tss"
	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain"
)

// Binance is a structure to sign and broadcast tx to binance chain used by signer mostly
type Binance struct {
	logger          zerolog.Logger
	cfg             config.ChainConfiguration
	chainID         string
	isTestNet       bool
	client          *http.Client
	accts           *BinanceMetaDataStore
	tssKeyManager   keys.KeyManager
	localKeyManager *keyManager
	thorchainBridge *thorclient.ThorchainBridge
	storage         *blockscanner.BlockScannerStorage
	blockScanner    *blockscanner.BlockScanner
	bnbScanner      *BinanceBlockScanner
}

// NewBinance create new instance of binance client
func NewBinance(thorKeys *thorclient.Keys, cfg config.ChainConfiguration, server *tssp.TssServer, thorchainBridge *thorclient.ThorchainBridge, m *metrics.Metrics) (*Binance, error) {
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

	b := &Binance{
		logger:          log.With().Str("module", "binance").Logger(),
		cfg:             cfg,
		accts:           NewBinanceMetaDataStore(),
		client:          &http.Client{},
		tssKeyManager:   tssKm,
		localKeyManager: localKm,
		thorchainBridge: thorchainBridge,
	}

	if err := b.checkIsTestNet(); err != nil {
		b.logger.Error().Err(err).Msg("fail to check if is testnet")
		return b, err
	}

	var path string // if not set later, will in memory storage
	if len(b.cfg.BlockScanner.DBPath) > 0 {
		path = fmt.Sprintf("%s/%s", b.cfg.BlockScanner.DBPath, b.cfg.BlockScanner.ChainID)
	}
	b.storage, err = blockscanner.NewBlockScannerStorage(path)
	if err != nil {
		return nil, fmt.Errorf("fail to create scan storage: %w", err)
	}

	b.bnbScanner, err = NewBinanceBlockScanner(b.cfg.BlockScanner, b.storage, b.isTestNet, m)
	if err != nil {
		return nil, fmt.Errorf("fail to create block scanner: %w", err)
	}

	b.blockScanner, err = blockscanner.NewBlockScanner(b.cfg.BlockScanner, b.storage, m, b.thorchainBridge, b.bnbScanner)
	if err != nil {
		return nil, fmt.Errorf("fail to create block scanner: %w", err)
	}

	return b, nil
}

func (b *Binance) Start(globalTxsQueue chan stypes.TxIn, globalErrataQueue chan stypes.ErrataBlock) {
	b.blockScanner.Start(globalTxsQueue)
}

func (b *Binance) Stop() {
	b.blockScanner.Stop()
}

func (b *Binance) GetConfig() config.ChainConfiguration {
	return b.cfg
}

// IsTestNet determinate whether we are running on test net by checking the status
func (b *Binance) checkIsTestNet() error {
	// Cached data after first call
	if b.isTestNet {
		return nil
	}

	u, err := url.Parse(b.cfg.RPCHost)
	if err != nil {
		return fmt.Errorf("Unable to parse rpc host: %s: %w", b.cfg.RPCHost, err)
	}

	u.Path = "/status"

	resp, err := b.client.Get(u.String())
	if err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			b.logger.Error().Err(err).Msg("fail to close resp body")
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
		return fmt.Errorf("fail to unmarshal body: %w", err)
	}

	b.chainID = status.Result.NodeInfo.Network
	b.isTestNet = b.chainID == "Binance-Chain-Nile"

	if b.isTestNet {
		types.Network = types.TestNetwork
	} else {
		types.Network = types.ProdNetwork
	}

	return nil
}

func (b *Binance) GetChain() common.Chain {
	return common.BNBChain
}

func (b *Binance) GetHeight() (int64, error) {
	u, err := url.Parse(b.cfg.RPCHost)
	if err != nil {
		return 0, fmt.Errorf("unable to parse dex host: %w", err)
	}
	u.Path = "abci_info"
	resp, err := b.client.Get(u.String())
	if err != nil {
		return 0, fmt.Errorf("fail to get request(%s): %w", u.String(), err) // fmt.Errorf("Get request failed: %w", err)
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

func (b *Binance) getGasFee(count uint64) common.Gas {
	coins := make(common.Coins, count)
	gasInfo := []sdk.Uint{
		sdk.NewUint(b.bnbScanner.singleFee),
		sdk.NewUint(b.bnbScanner.multiFee),
	}
	return common.CalcGasPrice(common.Tx{Coins: coins}, common.BNBAsset, gasInfo)
}

// SignTx sign the the given TxArrayItem
func (b *Binance) SignTx(tx stypes.TxOutItem, height int64) ([]byte, error) {
	var payload []msg.Transfer

	toAddr, err := types.AccAddressFromBech32(tx.ToAddress.String())
	if err != nil {
		return nil, fmt.Errorf("fail to parse account address(%s) :%w", tx.ToAddress.String(), err)
	}

	var gasCoin common.Coins

	// for yggdrasil, need to left some coin to pay for fee, this logic is per chain, given different chain charge fees differently
	if strings.EqualFold(tx.Memo, thorchain.NewYggdrasilReturn(height).String()) {
		gas := b.getGasFee(uint64(len(tx.Coins)))
		gasCoin = gas.ToCoins()
	}
	var coins types.Coins
	for _, coin := range tx.Coins {
		// deduct gas coin
		for _, gc := range gasCoin {
			if coin.Asset.Equals(gc.Asset) {
				coin.Amount = common.SafeSub(coin.Amount, gc.Amount)
			}
		}

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
	meta := b.accts.Get(tx.VaultPubKey)
	if currentHeight > meta.BlockHeight {
		acc, err := b.GetAccount(tx.VaultPubKey)
		if err != nil {
			return nil, fmt.Errorf("fail to get account info: %w", err)
		}
		meta = BinanceMetadata{
			AccountNumber: acc.AccountNumber,
			SeqNumber:     acc.Sequence,
			BlockHeight:   currentHeight,
		}
		b.accts.Set(tx.VaultPubKey, meta)
	}
	b.logger.Info().Int64("account_number", meta.AccountNumber).Int64("sequence_number", meta.SeqNumber).Msg("account info")
	signMsg := btx.StdSignMsg{
		ChainID:       b.chainID,
		Memo:          tx.Memo,
		Msgs:          []msg.Msg{sendMsg},
		Source:        btx.Source,
		Sequence:      meta.SeqNumber,
		AccountNumber: meta.AccountNumber,
	}
	rawBz, err := b.signMsg(signMsg, fromAddr, tx.VaultPubKey, height, tx)
	if err != nil {
		return nil, fmt.Errorf("fail to sign message: %w", err)
	}

	if len(rawBz) == 0 {
		// the transaction was already signed
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

// signMsg is design to sign a given message until it success or the same message had been send out by other signer
func (b *Binance) signMsg(signMsg btx.StdSignMsg, from string, poolPubKey common.PubKey, height int64, txOutItem stypes.TxOutItem) ([]byte, error) {
	keySignParty, err := b.thorchainBridge.GetKeysignParty(poolPubKey)
	if err != nil {
		b.logger.Error().Err(err).Msg("fail to get keysign party")
		return nil, err
	}
	rawBytes, err := b.sign(signMsg, poolPubKey, keySignParty)
	if err == nil && rawBytes != nil {
		return rawBytes, nil
	}
	var keysignError tss.KeysignError
	if errors.As(err, &keysignError) {
		if len(keysignError.Blame.BlameNodes) == 0 {
			// TSS doesn't know which node to blame
			return nil, err
		}

		// key sign error forward the keysign blame to thorchain
		txID, err := b.thorchainBridge.PostKeysignFailure(keysignError.Blame, height, txOutItem.Memo, txOutItem.Coins)
		if err != nil {
			b.logger.Error().Err(err).Msg("fail to post keysign failure to thorchain")
			return nil, err
		} else {
			b.logger.Info().Str("tx_id", txID.String()).Msgf("post keysign failure to thorchain")
			return nil, fmt.Errorf("sent keysign failure to thorchain")
		}
	}
	b.logger.Error().Err(err).Msgf("fail to sign msg with memo: %s", signMsg.Memo)
	// should THORNode give up? let's check the seq no on binance chain
	// keep in mind, when THORNode don't run our own binance full node, THORNode might get rate limited by binance
	return nil, err
}

func (b *Binance) GetAccount(pkey common.PubKey) (common.Account, error) {
	addr := b.GetAddress(pkey)
	address, err := types.AccAddressFromBech32(addr)
	if err != nil {
		b.logger.Error().Err(err).Msgf("fail to get parse address: %s", addr)
		return common.Account{}, err
	}
	u, err := url.Parse(b.cfg.RPCHost)
	if err != nil {
		log.Fatal().Msgf("Error parsing rpc (%s): %s", b.cfg.RPCHost, err)
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
func (b *Binance) BroadcastTx(tx stypes.TxOutItem, hexTx []byte) error {
	u, err := url.Parse(b.cfg.RPCHost)
	if err != nil {
		log.Error().Msgf("Error parsing rpc (%s): %s", b.cfg.RPCHost, err)
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
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("fail to read response body: %w", err)
	}

	// NOTE: we can actually see two different json responses for the same end.
	// This complicates things pretty well.
	// Sample 1: { "height": "0", "txhash": "D97E8A81417E293F5B28DDB53A4AD87B434CA30F51D683DA758ECC2168A7A005", "raw_log": "[{\"msg_index\":0,\"success\":true,\"log\":\"\",\"events\":[{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"set_observed_txout\"}]}]}]", "logs": [ { "msg_index": 0, "success": true, "log": "", "events": [ { "type": "message", "attributes": [ { "key": "action", "value": "set_observed_txout" } ] } ] } ] }
	// Sample 2: { "height": "0", "txhash": "6A9AA734374D567D1FFA794134A66D3BF614C4EE5DDF334F21A52A47C188A6A2", "code": 4, "raw_log": "{\"codespace\":\"sdk\",\"code\":4,\"message\":\"signature verification failed; verify correct account sequence and chain-id\"}" }
	var commit stypes.Commit
	err = json.Unmarshal(body, &commit)
	if err != nil || len(commit.Logs) == 0 {
		b.logger.Error().Err(err).Msgf("fail unmarshal commit: %s", string(body))

		var badCommit stypes.BadCommit // since commit doesn't work, lets try bad commit
		err = json.Unmarshal(body, &badCommit)
		if err != nil {
			b.logger.Error().Err(err).Msg("fail unmarshal bad commit")
			return fmt.Errorf("fail to unmarshal bad commit: %w", err)
		}

		// check for any failure logs
		// Error code 4 is used for bad account sequence number. We expect to
		// see this often because in TSS, multiple nodes will broadcast the
		// same sequence number but only one will be successful. We can just
		// drop and ignore in these scenarios. In 1of1 signing, we can also
		// drop and ignore. The reason being, thorchain will attempt to again
		// later.
		// Error code 5 is insufficient funds, ignore theses
		if badCommit.Code > 0 && badCommit.Code != int(sdk.CodeUnauthorized) && badCommit.Code != int(sdk.CodeInsufficientFunds) {
			err := errors.New(badCommit.Log)
			b.logger.Error().Err(err).Msg("fail to broadcast")
			return fmt.Errorf("fail to broadcast: %w", err)
		}
	}

	for _, log := range commit.Logs {
		if !log.Success {
			err := errors.New(log.Log)
			b.logger.Error().Err(err).Msg("fail to broadcast")
			return fmt.Errorf("fail to broadcast: %w", err)
		}
	}

	// increment sequence number
	b.accts.SeqInc(tx.VaultPubKey)

	return nil
}
