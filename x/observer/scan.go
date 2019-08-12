package observer

import (
	"fmt"
	"sort"
	"strconv"
	"net/url"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"github.com/go-redis/redis"
	log "github.com/rs/zerolog/log"

	"gitlab.com/thorchain/bepswap/observe/x/storage"
	types "gitlab.com/thorchain/bepswap/observe/x/observer/types"
)

type Scanner struct {
	Db *redis.Client
	PoolAddress string
	DexHost string
	RpcHost string
}

func NewScanner(poolAddress, dexHost, rpcHost string) *Scanner {
	return &Scanner{
		Db: storage.RedisClient(),
		PoolAddress: poolAddress,
		DexHost: dexHost,
		RpcHost: rpcHost,
	}
}

// Process blocks as they are placed into the channel. In order to 
// support multi-send, we need to query the RPC service from a given node.
func (s *Scanner) ProcessBlocks(blocks []int, scanChan chan []byte) {
	sort.Ints(blocks)

	min := int64(s.getLastBlock())
	if min == 0 {
		min = int64(blocks[0])
	}

	max := int64(blocks[len(blocks)-1])
	s.setLastBlock(max)

	for block := min; block <= max; block++ {
		log.Info().Msgf("Scanning block %v...", block)

		uri := url.URL{
			Scheme: "https",
			Host: s.RpcHost,
			Path: "tx_search",
		}

		q := uri.Query()
		q.Set("query", fmt.Sprintf("\"tx.height=%v\"", block))
		q.Set("prove", "true")
		uri.RawQuery = q.Encode()

		resp, _ := http.Get(uri.String())
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		var block types.Block
		json.Unmarshal(body, &block)

		var inTx types.InTx

		for _, txn := range block.Result.Txs {
			inTx.TxArray = append(inTx.TxArray, types.TxItem{Tx: txn.Hash})
			blockHeight, _ := strconv.ParseInt(txn.Height, 10, 64)
			inTx.BlockHeight = int(blockHeight)

			inTx = s.QueryTxn(inTx)
			json, _ := json.Marshal(inTx)
			scanChan <- json
		}
	}
}

// Call the REST API to get specific details of the transaction, and if it was for 
// our particular pool address.
func (s *Scanner) QueryTxn(inTx types.InTx) types.InTx {
	for _, txItem := range inTx.TxArray {
		uri := url.URL{
			Scheme: "https",
			Host: s.DexHost,
			Path: fmt.Sprintf("api/v1/tx/%v", txItem.Tx),
		}

		resp, _ := http.Get(uri.String())
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		var tx types.Tx
		json.Unmarshal(body, &tx)

		// @todo 	This is similar to what happens inside the socket logic - to keep
		// 				things DRY, suggest this be handled elsewhere.
		for _, msg := range tx.Tx.Value.Msg {
			for i, output := range msg.Value.Outputs {
				if output.Address == s.PoolAddress {
					sender := msg.Value.Inputs[i]

					for _, coin := range sender.Coins {
						parsedAmt, _ := strconv.ParseFloat(coin.Amount, 64)
						amount := parsedAmt*100000000

						txItem := types.TxItem{Tx: tx.Hash,
							Memo: tx.Tx.Value.Memo,
							Sender: sender.Address,
							Coins: types.Coins{
								Denom: coin.Denom,
								Amount: fmt.Sprintf("%.0f", amount),
							},
						}
						inTx.TxArray = append(inTx.TxArray, txItem)
					}				
				}
			}
		}
	}

	return inTx
}

func (s *Scanner) getLastBlock() int64 {
	data, _ := s.Db.Get("lastBlock").Result()
	blockHeight, _ := strconv.ParseInt(data, 10, 64)

	return blockHeight
}

func (s *Scanner) setLastBlock(blockHeight int64) {
	err := s.Db.Set("lastBlock", blockHeight, 0).Err()
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
	}
}
