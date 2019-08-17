package observer

import (
	"encoding/json"
	"strconv"

	log "github.com/rs/zerolog/log"
	"github.com/syndtr/goleveldb/leveldb"

	ctypes "gitlab.com/thorchain/bepswap/observe/common/types"
	"gitlab.com/thorchain/bepswap/observe/x/statechain"
	"gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

type Observer struct {
	Db           *leveldb.DB
	WebSocket    *WebSocket
	BlockScan    *BlockScan
	SocketTxChan chan []byte
	BlockTxChan  chan []byte
}

func NewObserver() *Observer {
	var db, _ = leveldb.OpenFile(ctypes.ObserverDbPath, nil)
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

		//log.Info().Msgf("Processing Transaction: %v", txIn)
		signed := statechain.Sign(txIn)
		go statechain.Send(signed)
	}
}

func (o *Observer) SavePos(block int64) {
	go func() {
		_ = o.Db.Put([]byte(strconv.FormatInt(block, 10)), []byte(strconv.FormatInt(1, 10)), nil)
	}()
}
