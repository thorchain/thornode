package observer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/rs/zerolog/log"

	"gitlab.com/thorchain/bepswap/common"
	ctypes "gitlab.com/thorchain/bepswap/observe/common/types"
	"gitlab.com/thorchain/bepswap/observe/x/binance"
	btypes "gitlab.com/thorchain/bepswap/observe/x/binance/types"
	stypes "gitlab.com/thorchain/bepswap/observe/x/statechain/types"
)

type WebSocket struct {
	TxInChan   chan []byte
	SocketChan chan []byte
	Binance    *binance.Binance
}

func NewWebSocket(txChan chan []byte) *WebSocket {
	binance := binance.NewBinance()
	return &WebSocket{
		TxInChan:   txChan,
		SocketChan: make(chan []byte),
		Binance:    binance,
	}
}

func (w *WebSocket) Start() {
	conn, err := w.Connect()
	if err != nil {
		log.Fatal().Msgf("There was an error while starting: %v", err)
	}

	log.Info().Msgf("Setting a keepalive of %v", ctypes.SocketPong)
	w.SetKeepAlive(conn)

	log.Info().Msg("Listening for events....")
	w.ParseMessage()
	w.ReadSocket(conn)
}

func (w *WebSocket) Connect() (*websocket.Conn, error) {
	url := url.URL{
		Scheme: "wss",
		Host:   ctypes.DEXHost,
		Path:   fmt.Sprintf("/api/ws/%s", ctypes.PoolAddress),
	}

	log.Info().Msgf("Opening up a connection to: %v", url.String())

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (w *WebSocket) SetKeepAlive(conn *websocket.Conn) {
	lastResponse := time.Now()
	conn.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		for {
			err := conn.WriteMessage(websocket.PingMessage, []byte("pong"))
			if err != nil {
				return
			}

			time.Sleep(ctypes.SocketPong / 2)
			if time.Now().Sub(lastResponse) > ctypes.SocketPong {
				conn.Close()
				return
			}
		}
	}()
}

func (w *WebSocket) ReadSocket(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			// @todo Reconnect if this fails.
			log.Error().Msgf("Read error: %s", err)
		}
		w.SocketChan <- message
	}
}

func (w *WebSocket) ParseMessage() {
	go func() {
		for {
			payload := <-w.SocketChan
			var txfr btypes.SocketTxfr

			err := json.Unmarshal(payload, &txfr)
			if err != nil {
				log.Error().Msgf("There was an error while parsing the event: %v", err)
			}

			if txfr.Stream == "transfers" {
				var txIn stypes.TxIn

				if txfr.Data.FromAddr != ctypes.PoolAddress {
					for _, txn := range txfr.Data.T {
						txItem := stypes.TxInItem{Tx: txfr.Data.Hash,
							Memo:   txfr.Data.Memo,
							Sender: txfr.Data.FromAddr,
						}

						for _, coin := range txn.Coins {
							parsedAmt, _ := strconv.ParseFloat(coin.Amount, 64)
							amount := parsedAmt * 100000000

							var token common.Coin
							token.Denom = common.Ticker(coin.Asset)
							token.Amount = common.Amount(fmt.Sprintf("%.0f", amount))
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

				txIn.BlockHeight = strconv.Itoa(txfr.Data.EventHeight)
				txIn.Count = strconv.Itoa(len(txIn.TxArray))

				json, err := json.Marshal(txIn)
				if err != nil {
					log.Error().Msgf("Error: %v", err)
				}

				w.TxInChan <- json
			}
		}
	}()
}
