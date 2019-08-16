package observer

import (
	"encoding/json"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/syndtr/goleveldb/leveldb"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
	config "gitlab.com/thorchain/bepswap/observe/config"

	"gitlab.com/thorchain/bepswap/observe/x/statechain"
	"gitlab.com/thorchain/bepswap/observe/x/statechain/types"
	stypes "gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
)

type Observer struct {
	Db           *leveldb.DB
	WebSocket    *WebSocket
	BlockScan    *BlockScan
	SocketTxChan chan []byte
	BlockTxChan  chan []byte
}

func NewObserver() *Observer {
	var db, _ = leveldb.OpenFile(config.ObserverDbPath, nil)
	socketTxChan := make(chan []byte)
	blockTxChan := make(chan []byte)

	return &Observer{
		Db:           db,
		WebSocket:    NewWebSocket(socketTxChan),
		BlockScan:    NewBlockScan(blockTxChan),
		SocketTxChan: socketTxChan,
		BlockTxChan:  blockTxChan,
	}
}

func (o *Observer) Start() {
	go o.ProcessTxnIn(o.BlockTxChan)
	go o.ProcessTxnIn(o.SocketTxChan)
	go o.BlockScan.Start()
	o.WebSocket.Start()
}

func (o *Observer) ProcessTxnIn(ch chan []byte) {
	for {
		var txIn types.TxIn
		payload := <-ch

		err := json.Unmarshal(payload, &txIn)
		if err != nil {
			log.Error().Msgf("Error: %v", err)
		}

		mode, _ := types.NewMode("sync")

		addr, err := sdk.AccAddressFromBech32(config.RuneAddress)
		if err != nil {
			log.Error().Msgf("Error: %v", err)
		}

		txs := make([]stypes.TxIn, len(txIn.TxArray))
		for i, item := range txIn.TxArray {
			// TODO: don't ignore this error
			txID, _ := common.NewTxID(item.Tx)
			bnbAddr, _ := common.NewBnbAddress(item.Sender)
			txs[i] = stypes.NewTxIn(
				txID,
				item.Coins,
				item.Memo,
				bnbAddr,
			)
		}

		signed, _ := statechain.Sign(txs, addr)
		go statechain.Send(signed, mode)
	}
}

func (o *Observer) SavePos(block int64) {
	go func() {
		_ = o.Db.Put([]byte(strconv.FormatInt(block, 10)), []byte(strconv.FormatInt(1, 10)), nil)
	}()
}
