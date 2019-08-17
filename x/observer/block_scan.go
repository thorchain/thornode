package observer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	log "github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	"gitlab.com/thorchain/bepswap/common"
	config "gitlab.com/thorchain/bepswap/observe/config"
	btypes "gitlab.com/thorchain/bepswap/observe/x/binance/types"
	stypes "gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

type BlockScan struct {
	TxInChan      chan []byte
	ScanChan      chan int64
	PreviousBlock int64
}

func NewBlockScan(txInChan chan []byte) *BlockScan {
	scanChan := make(chan int64)
	return &BlockScan{
		TxInChan:      txInChan,
		ScanChan:      scanChan,
		PreviousBlock: 0,
	}
}

func (b *BlockScan) Start() {
	b.TxSearch()
	b.ScanBlocks()
}

func (b *BlockScan) ScanBlocks() {
	go func() {
		for {
			uri := url.URL{
				Scheme: "https",
				Host:   config.RPCHost,
				Path:   "block",
			}

			req := fasthttp.AcquireRequest()
			req.SetRequestURI(uri.String())

			resp := fasthttp.AcquireResponse()
			client := &fasthttp.Client{}
			_ = client.Do(req, resp)

			body := resp.Body()
			var tx btypes.RPCBlock
			_ = json.Unmarshal(body, &tx)

			block := tx.Result.Block.Header.Height
			parsedBlock, _ := strconv.ParseInt(block, 10, 64)

			if b.PreviousBlock != parsedBlock {
				log.Info().Msgf("Processing Binance Block Height: %v", parsedBlock)
				b.PreviousBlock = parsedBlock
				b.ScanChan <- parsedBlock
			}
		}
	}()
}

func (b *BlockScan) TxSearch() {
	go func() {
		for {
			block := <-b.ScanChan

			uri := url.URL{
				Scheme: "https",
				Host:   config.RPCHost,
				Path:   "tx_search",
			}

			q := uri.Query()
			q.Set("query", fmt.Sprintf("\"tx.height=%v\"", block))
			q.Set("prove", "true")
			uri.RawQuery = q.Encode()

			req := fasthttp.AcquireRequest()
			req.SetRequestURI(uri.String())

			resp := fasthttp.AcquireResponse()
			client := &fasthttp.Client{}
			_ = client.Do(req, resp)

			body := resp.Body()
			var query btypes.RPCTxSearch
			_ = json.Unmarshal(body, &query)

			var txIn stypes.TxIn
			for _, txn := range query.Result.Txs {
				txIn.TxArray = append(txIn.TxArray, stypes.TxInItem{Tx: txn.Hash})
				txIn = b.QueryTx(txIn)
			}

			txIn.BlockHeight = strconv.FormatInt(block, 10)
			txIn.Count = strconv.Itoa(len(txIn.TxArray))

			json, _ := json.Marshal(txIn)
			if len(txIn.TxArray) >= 1 {
				log.Info().Msgf("%v", string(json))
			}

			b.TxInChan <- json
		}
	}()
}

func (b *BlockScan) QueryTx(txIn stypes.TxIn) stypes.TxIn {
	for i, txItem := range txIn.TxArray {
		uri := url.URL{
			Scheme: "https",
			Host:   config.RPCHost,
			Path:   fmt.Sprintf("api/v1/tx/%v", txItem.Tx),
		}

		q := uri.Query()
		q.Set("format", "json")
		uri.RawQuery = q.Encode()

		req := fasthttp.AcquireRequest()
		req.SetRequestURI(uri.String())

		resp := fasthttp.AcquireResponse()
		client := &fasthttp.Client{}
		_ = client.Do(req, resp)

		body := resp.Body()

		var tx btypes.ApiTx
		_ = json.Unmarshal(body, &tx)

		for _, msg := range tx.Tx.Value.Msg {
			for j, output := range msg.Value.Outputs {
				if output.Address == config.PoolAddress {
					sender := msg.Value.Inputs[j]

					for _, coin := range sender.Coins {
						parsedAmt, _ := strconv.ParseFloat(coin.Amount, 64)
						amount := parsedAmt * 100000000

						txIn.TxArray[i].Memo = tx.Tx.Value.Memo
						txIn.TxArray[i].Sender = sender.Address
						token := common.Coin{
							Denom:  common.Ticker(coin.Denom),
							Amount: common.NewAmountFromFloat(amount),
						}
						txIn.TxArray[i].Coins = append(txIn.TxArray[i].Coins, token)
					}
				}
			}

			for _, input := range msg.Value.Inputs {
				if input.Address == config.PoolAddress {
					txIn.TxArray[i].Memo = tx.Tx.Value.Memo
				}
			}
		}
	}

	return txIn
}
