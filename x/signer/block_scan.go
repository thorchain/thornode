package signer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	log "github.com/rs/zerolog/log"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/valyala/fasthttp"

	"gitlab.com/thorchain/bepswap/common"
	ctypes "gitlab.com/thorchain/bepswap/observe/common/types"
	stypes "gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

type BlockScan struct {
	Db        *leveldb.DB
	TxOutChan chan []byte
}

func NewBlockScan(db *leveldb.DB, txOutChan chan []byte) *BlockScan {
	return &BlockScan{
		Db:        db,
		TxOutChan: txOutChan,
	}
}

func (b *BlockScan) Start() {
	for {
		<-time.After(3 * time.Second)
		blockHeight := b.GetLastBlock()
		log.Info().Msgf("Processing Statechain Block Height: %v", blockHeight+1)

		b.SetLastBlock(blockHeight + 1)
		go b.ScanBlock(blockHeight)
	}
}

func (b *BlockScan) ScanBlock(blockHeight int64) {
	uri := url.URL{
		Scheme: "http",
		Host:   ctypes.ChainHost,
		Path:   fmt.Sprintf("/swapservice/txoutarray/%v", blockHeight),
	}

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(uri.String())

	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	_ = client.Do(req, resp)

	body := resp.Body()

	var txOut stypes.TxOut
	err := json.Unmarshal(body, &txOut)
	if err != nil {
		log.Error().Msgf("Error: %v", err)
	}

	if len(txOut.TxArray) >= 1 {
		for i, txArr := range txOut.TxArray {
			for j, coin := range txArr.Coins {
				amt := coin.Amount.Float64()
				txOut.TxArray[i].Coins[j].Amount = common.Amount(fmt.Sprintf("%.0f", amt))
			}
		}

		json, _ := json.Marshal(txOut)
		b.TxOutChan <- json
	}
}

func (b *BlockScan) GetLastBlock() int64 {
	data, err := b.Db.Get([]byte("LAST_BLOCK"), nil)
	if err != nil {
		log.Error().Msgf("Error: %v", err)
	}

	blockHeight, _ := strconv.ParseInt(string(data), 10, 64)

	return blockHeight
}

func (b *BlockScan) SetLastBlock(blockHeight int64) {
	_ = b.Db.Put([]byte("LAST_BLOCK"), []byte(strconv.FormatInt(blockHeight, 10)), nil)
	_ = b.Db.Put([]byte(strconv.FormatInt(blockHeight, 10)), []byte(strconv.FormatInt(1, 10)), nil)
}
