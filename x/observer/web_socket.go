package observer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/thorchain/bepswap/common"

	"gitlab.com/thorchain/bepswap/observe/config"
	btypes "gitlab.com/thorchain/bepswap/observe/x/binance/types"
	stypes "gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

// WebSocket struct
type WebSocket struct {
	wg            *sync.WaitGroup
	cfg           config.Configuration
	logger        zerolog.Logger
	txInChan      chan stypes.TxIn
	socketChan    chan []byte
	websocketConn *websocket.Conn
	stopChan      chan struct{}
}

// NewWebSocket create a new instance of web socket
func NewWebSocket(cfg config.Configuration) (*WebSocket, error) {
	if len(cfg.DEXHost) == 0 {
		return nil, errors.New("DEXHost is empty")
	}
	if len(cfg.PoolAddress) == 0 {
		return nil, errors.New("pool address is empty")
	}
	return &WebSocket{
		cfg:        cfg,
		logger:     log.Logger.With().Str("module", "websocket").Logger(),
		txInChan:   make(chan stypes.TxIn, cfg.MessageProcessor),
		socketChan: make(chan []byte, cfg.MessageProcessor),
		stopChan:   make(chan struct{}),
		wg:         &sync.WaitGroup{},
	}, nil
}

func (w *WebSocket) Start() error {
	log.Debug().Msg("Listening for events....")
	for idx := 1; idx <= w.cfg.MessageProcessor; idx++ {
		w.wg.Add(1)
		go w.parseMessage(idx)
	}
	w.wg.Add(1)
	go w.websocketManager()
	return nil
}

func (w *WebSocket) connect() (*websocket.Conn, error) {
	connectURL := url.URL{
		Scheme: "wss",
		Host:   w.cfg.DEXHost,
		Path:   fmt.Sprintf("/api/ws/%s", w.cfg.PoolAddress),
	}
	w.logger.Info().Msgf("opening up a connection to %s", connectURL.String())

	// TODO maybe check the response
	conn, _, err := websocket.DefaultDialer.Dial(connectURL.String(), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to dial url %s", connectURL.String())
	}
	return conn, nil
}

func (w *WebSocket) Stop() error {
	close(w.stopChan)
	if w.websocketConn != nil {
		if err := w.websocketConn.Close(); nil != err {
			w.logger.Error().Err(err).Msg("fail to close websocket")
		}
	}
	w.wg.Wait()
	close(w.txInChan)
	return nil
}
func (w *WebSocket) GetMessages() <-chan stypes.TxIn {
	return w.txInChan
}

func (w *WebSocket) websocketManager() {
	defer w.wg.Done()
	for {
		select {
		case <-w.stopChan:
			return
		default:
			// logic need to handle socket get disconnected
			conn, err := w.connect()
			if err != nil {
				w.logger.Err(err).Msg("fail to connec to websocket")
				// TODO implement exponential backoff and retry here , right now , let's keep trying forever
				continue
			}
			w.websocketConn = conn
			w.readFromSocket(conn)
		}
	}

}
func (w *WebSocket) readFromSocket(conn *websocket.Conn) {
	w.logger.Debug().Msg("start read from websocket")
	defer func() {
		if err := conn.Close(); nil != err {
			w.logger.Error().Err(err).Msg("fail to close websocket")
		}
		w.logger.Debug().Msg("stop read from websocket")
	}()
	for {
		select {
		case <-w.stopChan:
			return
		default:
			w.logger.Debug().Msg("time to read messages")

			msgType, message, err := conn.ReadMessage()
			if nil != err {
				w.logger.Error().Err(err).Msg("permanent error,socket closed")
				return
			}
			w.logger.Debug().Int("msgType", msgType).Err(err).Msg(string(message))
			switch msgType {
			// Ping And Pong message will never get to here, but I leave it here just want to be sure.
			// TODO to be removed a bit later
			case websocket.PingMessage:
				w.logger.Debug().Msg("get ping msg")
			case websocket.PongMessage:
				w.logger.Debug().Msg("get pong msg")
			case websocket.TextMessage:
				w.logger.Info().Msg("get text message")
				w.socketChan <- message
			default:
				// other message type means error
				w.logger.Debug().Int("msgtype", msgType).Msg("unexpected message type")
				return
			}
		}

	}
}

func (w *WebSocket) processMessage(payload []byte) error {
	if len(payload) == 0 {
		return errors.New("message payload is empty")
	}
	var txfr btypes.SocketTxfr
	err := json.Unmarshal(payload, &txfr)
	if err != nil {
		return errors.Wrap(err, "fail to parse the event")
	}

	if strings.EqualFold(txfr.Stream, "transfers") {
		var txIn stypes.TxIn
		if txfr.Data.FromAddr != w.cfg.PoolAddress.String() {
			for _, txn := range txfr.Data.T {
				txItem := stypes.TxInItem{Tx: txfr.Data.Hash,
					Memo:   txfr.Data.Memo,
					Sender: txfr.Data.FromAddr,
				}

				for _, coin := range txn.Coins {
					parsedAmt, _ := strconv.ParseFloat(coin.Amount, 64)
					amount := common.FloatToUint(parsedAmt)

					var token common.Coin
					token.Denom = common.Ticker(coin.Asset)
					token.Amount = amount
					txItem.Coins = append(txItem.Coins, token)
				}

				txIn.TxArray = append(txIn.TxArray, txItem)
			}
		} else {
			txItem := stypes.TxInItem{Tx: txfr.Data.Hash,
				Memo:   txfr.Data.Memo,
				Sender: txfr.Data.FromAddr,
			}
			txIn.TxArray = append(txIn.TxArray, txItem)
		}
		// TODO change the types
		txIn.BlockHeight = strconv.Itoa(txfr.Data.EventHeight)
		txIn.Count = strconv.Itoa(len(txIn.TxArray))

		w.txInChan <- txIn
	}
	return nil
}

func (w *WebSocket) parseMessage(idx int) {
	w.logger.Info().Int("idx", idx).Msg("start to process messages from websocket")
	defer w.logger.Info().Int("idx", idx).Msg("stopped process messages from websocket")
	defer w.wg.Done()
	for {
		select {
		case <-w.stopChan:
			// we need to drop everything and stop processing messages
			return
		case payload := <-w.socketChan:
			if err := w.processMessage(payload); nil != err {
				w.logger.Error().Err(err).Str("msg", string(payload)).Msg("fail to process a message")
			}
		}
	}

}
