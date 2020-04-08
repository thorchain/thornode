package blockscanner

// This implementation is design for cosmos based blockchains

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/binance-chain/go-sdk/common/types"
	bmsg "github.com/binance-chain/go-sdk/types/msg"
	"github.com/binance-chain/go-sdk/types/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/go-amino"

	"gitlab.com/thorchain/thornode/bifrost/config"
	"gitlab.com/thorchain/thornode/bifrost/pubkeymanager"
	stypes "gitlab.com/thorchain/thornode/bifrost/thorclient/types"
	"gitlab.com/thorchain/thornode/common"
)

type QueryResult struct {
	Result struct {
		Response struct {
			Value string `json:"value"`
		} `json:"response"`
	} `json:"result"`
}

type itemData struct {
	Txs []string `json:"txs"`
}

type itemHeader struct {
	Height string `json:"height"`
}

type itemBlock struct {
	Header itemHeader `json:"header"`
	Data   itemData   `json:"data"`
}

type itemResult struct {
	Block itemBlock `json:"block"`
}

type item struct {
	Jsonrpc string     `json:"jsonrpc"`
	ID      string     `json:"id"`
	Result  itemResult `json:"result"`
}

type CosmosSupplemental struct {
	logger     zerolog.Logger
	cfg        config.BlockScannerConfiguration
	rpcHost    string
	httpClient *http.Client
	pubkeyMgr  pubkeymanager.PubKeyValidator
	singleFee  uint64
	multiFee   uint64
}

func NewCosmosSupplemental(cfg config.BlockScannerConfiguration, pkmgr pubkeymanager.PubKeyValidator) CosmosSupplemental {

	rpcHost := cfg.RPCHost
	if !strings.HasPrefix(rpcHost, "http") {
		rpcHost = fmt.Sprintf("http://%s", rpcHost)
	}

	return CosmosSupplemental{
		logger:  log.Logger.With().Str("module", cfg.ChainID.String()).Logger(),
		cfg:     cfg,
		rpcHost: rpcHost,
		httpClient: &http.Client{
			Timeout: cfg.HttpRequestTimeout,
		},
		pubkeyMgr: pkmgr,
	}
}

func (cms CosmosSupplemental) GetTxs(height int64) (stypes.TxIn, bool, error) {
	noTxs := stypes.TxIn{}
	url := cms.getURL(height)
	bz, err := cms.getFromHttp(url)
	if err != nil {
		cms.logger.Error().Err(err).Msgf("fail to get block data: %s", url)
		return noTxs, false, err
	}

	rawTxs, err := cms.UnmarshalBlock(bz)
	if err != nil {
		cms.logger.Error().Err(err).Msgf("fail to unmarshal block data: %s", url)
		return noTxs, false, err
	}

	txIn, err := cms.processBlock(height, rawTxs)
	return txIn, true, err
}

func (cms CosmosSupplemental) getURL(height int64) string {
	u, _ := url.Parse(cms.rpcHost)
	u.Path = "block"
	if height > 0 {
		u.RawQuery = fmt.Sprintf("height=%d", height)
	}
	return u.String()
}

func (cms CosmosSupplemental) getFromHttp(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create http request")
	}
	resp, err := cms.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to get from %s ", url)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			cms.logger.Error().Err(err).Msg("fail to close http response body.")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code:%d from %s", resp.StatusCode, url)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// test if our response body is an error block json format
	errorBlock := struct {
		Error struct {
			Code    int64  `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		} `json:"error"`
	}{}

	_ = json.Unmarshal(buf, &errorBlock) // ignore error
	if errorBlock.Error.Code != 0 {
		return nil, fmt.Errorf(
			"%s (%d): %s",
			errorBlock.Error.Message,
			errorBlock.Error.Code,
			errorBlock.Error.Data,
		)
	}

	return buf, nil
}

func (cosmos CosmosSupplemental) UnmarshalBlock(buf []byte) ([]string, error) {
	var block item
	err := json.Unmarshal(buf, &block)
	if err != nil {
		return nil, errors.Wrap(err, "fail to unmarshal body to rpcBlock")
	}

	// sanity check to confirm we have some real data
	if len(block.Result.Block.Header.Height) == 0 {
		return nil, errors.Wrap(err, "fail to get data: missing data fields")
	}

	return block.Result.Block.Data.Txs, nil
}

func (cms CosmosSupplemental) processBlock(height int64, rawTxs []string) (stypes.TxIn, error) {
	noTxs := stypes.TxIn{}
	cms.logger.Debug().Int64("block", height).Int("txs", len(rawTxs)).Msg("txs")
	if len(rawTxs) == 0 {
		return noTxs, nil
	}

	// update our gas fees from binance RPC node
	if err := cms.updateFees(height); err != nil {
		cms.logger.Error().Err(err).Msg("fail to update Binance gas fees")
	}

	// TODO implement pagination appropriately
	var txIn stypes.TxIn
	for _, txn := range rawTxs {
		hash, err := cms.getTxHash(txn)
		if err != nil {
			cms.logger.Error().Err(err).Str("tx", txn).Msg("fail to get tx hash from raw data")
			return noTxs, errors.Wrap(err, "fail to get tx hash from tx raw data")
		}

		txItemIns, err := cms.fromTxToTxIn(hash, txn)
		if err != nil {
			cms.logger.Error().Err(err).Str("hash", hash).Msg("fail to get one tx from server")
			// if THORNode fail to get one tx hash from server, then THORNode should bail, because THORNode might miss tx
			// if THORNode bail here, then THORNode should retry later
			return noTxs, errors.Wrap(err, "fail to get one tx from server")
		}
		if len(txItemIns) > 0 {
			txIn.TxArray = append(txIn.TxArray, txItemIns...)
			cms.logger.Info().Str("hash", hash).Msg("THORNode got one tx")
		}
	}
	if len(txIn.TxArray) == 0 {
		cms.logger.Debug().Int64("block", height).Msg("no tx need to be processed in this block")
		return noTxs, nil
	}

	txIn.BlockHeight = strconv.FormatInt(height, 10)
	txIn.Count = strconv.Itoa(len(txIn.TxArray))
	txIn.Chain = common.BNBChain
	return txIn, nil
}

func (cms CosmosSupplemental) getCoinsForTxIn(outputs []bmsg.Output) (common.Coins, error) {
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

// getTxHash return hex formatted value of tx hash
// raw tx base 64 encoded -> base64 decode -> sha256sum = tx hash
func (cms CosmosSupplemental) getTxHash(encodedTx string) (string, error) {
	decodedTx, err := base64.StdEncoding.DecodeString(encodedTx)
	if err != nil {
		return "", errors.Wrap(err, "fail to decode tx")
	}
	return fmt.Sprintf("%X", sha256.Sum256(decodedTx)), nil
}

func (cms CosmosSupplemental) fromTxToTxIn(hash, encodedTx string) ([]stypes.TxInItem, error) {
	if len(encodedTx) == 0 {
		return nil, errors.New("tx is empty")
	}
	buf, err := base64.StdEncoding.DecodeString(encodedTx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to decode tx")
	}
	var t tx.StdTx
	if err := tx.Cdc.UnmarshalBinaryLengthPrefixed(buf, &t); err != nil {
		return nil, errors.Wrap(err, "fail to unmarshal tx.StdTx")
	}

	return cms.fromStdTx(hash, t)
}

// fromStdTx - process a stdTx
func (cms *CosmosSupplemental) fromStdTx(hash string, stdTx tx.StdTx) ([]stypes.TxInItem, error) {
	var err error
	var txs []stypes.TxInItem

	// TODO: It is also possible to have multiple inputs/outputs within a
	// single stdTx, which THORNode are not yet accounting for.
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
			txInItem.Coins, err = cms.getCoinsForTxIn(sendMsg.Outputs)
			if err != nil {
				return nil, errors.Wrap(err, "fail to convert coins")
			}

			// Calculate gas for this tx
			txInItem.Gas = common.CalcGasPrice(common.Tx{Coins: txInItem.Coins}, common.BNBAsset, []sdk.Uint{sdk.NewUint(cms.singleFee), sdk.NewUint(cms.multiFee)})

			if ok := cms.MatchedAddress(txInItem); !ok {
				continue
			}

			// NOTE: the following could result in the same tx being added
			// twice, which is expected. We want to make sure we generate both
			// a inbound and outbound txn, if we both apply.

			// check if the from address is a valid pool
			if ok, cpi := cms.pubkeyMgr.IsValidPoolAddress(txInItem.Sender, common.BNBChain); ok {
				txInItem.ObservedPoolAddress = cpi.PubKey.String()
				txs = append(txs, txInItem)
			}
			// check if the to address is a valid pool address
			if ok, cpi := cms.pubkeyMgr.IsValidPoolAddress(txInItem.To, common.BNBChain); ok {
				txInItem.ObservedPoolAddress = cpi.PubKey.String()
				txs = append(txs, txInItem)
			} else {
				// Apparently we don't recognize where we are sending funds to.
				// Lets check if we should because its an internal transaction
				// moving funds between vaults (for example). If it is, lets
				// manually trigger an update of pubkeys, then check again...
				switch strings.ToLower(txInItem.Memo) {
				case "migrate", "yggdrasil-", "yggdrasil+":
					cms.pubkeyMgr.FetchPubKeys()
					if ok, cpi := cms.pubkeyMgr.IsValidPoolAddress(txInItem.To, common.BNBChain); ok {
						txInItem.ObservedPoolAddress = cpi.PubKey.String()
						txs = append(txs, txInItem)
					}
				}
			}

		default:
			continue
		}
	}
	return txs, nil
}

func (cms CosmosSupplemental) updateFees(height int64) error {
	url := fmt.Sprintf("%s/abci_query?path=\"/param/fees\"&height=%d", cms.rpcHost, height)
	resp, err := cms.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get current gas fees: non 200 error (%d)", resp.StatusCode)
	}

	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var result QueryResult
	if err := json.Unmarshal(bz, &result); err != nil {
		return err
	}

	val, err := base64.StdEncoding.DecodeString(result.Result.Response.Value)
	if err != nil {
		return err
	}

	var fees []types.FeeParam
	cdc := amino.NewCodec()
	types.RegisterWire(cdc)
	err = cdc.UnmarshalBinaryLengthPrefixed(val, &fees)
	if err != nil {
		return err
	}

	for _, fee := range fees {
		if fee.GetParamType() == types.TransferFeeType {
			if err := fee.Check(); err != nil {
				return err
			}

			transferFee := fee.(*types.TransferFeeParam)
			if transferFee.FixedFeeParams.Fee > 0 {
				cms.singleFee = uint64(transferFee.FixedFeeParams.Fee)
			}
			if transferFee.MultiTransferFee > 0 {
				cms.multiFee = uint64(transferFee.MultiTransferFee)
			}
		}
	}

	return nil
}

func (cms CosmosSupplemental) MatchedAddress(txInItem stypes.TxInItem) bool {
	// Check if we are migrating our funds...
	if ok := cms.isMigration(txInItem.Sender, txInItem.Memo); ok {
		cms.logger.Debug().Str("memo", txInItem.Memo).Msg("migrate")
		return true
	}

	// Check if our pool is registering a new yggdrasil pool. Ie
	// sending the staked assets to the user
	if ok := cms.isRegisterYggdrasil(txInItem.Sender, txInItem.Memo); ok {
		cms.logger.Debug().Str("memo", txInItem.Memo).Msg("yggdrasil+")
		return true
	}

	// Check if out pool is de registering a yggdrasil pool. Ie sending
	// the bond back to the user
	if ok := cms.isDeregisterYggdrasil(txInItem.Sender, txInItem.Memo); ok {
		cms.logger.Debug().Str("memo", txInItem.Memo).Msg("yggdrasil-")
		return true
	}

	// Check if THORNode are sending from a yggdrasil address
	if ok := cms.isYggdrasil(txInItem.Sender); ok {
		cms.logger.Debug().Str("assets sent from yggdrasil pool", txInItem.Memo).Msg("fill order")
		return true
	}

	// Check if THORNode are sending to a yggdrasil address
	if ok := cms.isYggdrasil(txInItem.To); ok {
		cms.logger.Debug().Str("assets to yggdrasil pool", txInItem.Memo).Msg("refill")
		return true
	}

	// outbound message from pool, when it is outbound, it does not matter how much coins THORNode send to customer for now
	if ok := cms.isOutboundMsg(txInItem.Sender, txInItem.Memo); ok {
		cms.logger.Debug().Str("memo", txInItem.Memo).Msg("outbound")
		return true
	}

	return false
}

// Check if memo is for registering an Asgard vault
func (cms CosmosSupplemental) isMigration(addr, memo string) bool {
	return cms.isAddrWithMemo(addr, memo, "migrate")
}

// Check if memo is for registering a Yggdrasil vault
func (cms CosmosSupplemental) isRegisterYggdrasil(addr, memo string) bool {
	return cms.isAddrWithMemo(addr, memo, "yggdrasil+")
}

// Check if memo is for de registering a Yggdrasil vault
func (cms CosmosSupplemental) isDeregisterYggdrasil(addr, memo string) bool {
	return cms.isAddrWithMemo(addr, memo, "yggdrasil-")
}

// Check if THORNode have an outbound yggdrasil transaction
func (cms CosmosSupplemental) isYggdrasil(addr string) bool {
	ok, _ := cms.pubkeyMgr.IsValidPoolAddress(addr, common.BNBChain)
	return ok
}

func (cms CosmosSupplemental) isOutboundMsg(addr, memo string) bool {
	return cms.isAddrWithMemo(addr, memo, "outbound")
}

func (cms CosmosSupplemental) isAddrWithMemo(addr, memo, targetMemo string) bool {
	match, _ := cms.pubkeyMgr.IsValidPoolAddress(addr, common.BNBChain)
	if !match {
		return false
	}
	lowerMemo := strings.ToLower(memo)
	if strings.HasPrefix(lowerMemo, targetMemo) {
		return true
	}
	return false
}
