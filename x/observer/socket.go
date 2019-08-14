package observer

import (
	"os"
	"fmt"
	"time"
	"strconv"
	"net/url"
	"encoding/json"

	"github.com/gorilla/websocket"
	log "github.com/rs/zerolog/log"

	"gitlab.com/thorchain/bepswap/observe/x/binance"
	ctypes "gitlab.com/thorchain/bepswap/observe/common/types"
	types "gitlab.com/thorchain/bepswap/observe/x/observer/types"
)

type Socket struct {
	Binance *binance.Binance
	PongWait time.Duration
}

func NewSocket() *Socket {
	binance := binance.NewBinance()
	return &Socket{
		Binance: binance,
		PongWait: 30*time.Second,
	}
}

func (s Socket) Start(conChan chan []byte) {
	conn, err := s.Connect()
	if err != nil {
		log.Fatal().Msgf("There was an error while starting: %v", err)
	}

	log.Info().Msgf("Setting a keepalive of %v", s.PongWait)
	s.SetKeepAlive(conn)

	ch := make(chan []byte)

	log.Info().Msg("%s Listening for events....")
	s.Process(ch, conChan)
	s.Read(ch, conn)
}

func (s Socket) Connect() (*websocket.Conn, error) {
	path := fmt.Sprintf("/api/ws/%s", ctypes.PoolAddress)
	url := url.URL{
		Scheme: "wss",
		Host: ctypes.DEXHost,
		Path: path,
	}

	log.Info().Msgf("Opening up a connection to: %v", url.String())

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (s Socket) SetKeepAlive(conn *websocket.Conn) {
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

			time.Sleep(s.PongWait / 2)
			if time.Now().Sub(lastResponse) > s.PongWait {
				conn.Close()
				return
			}
		}
	}()
}

func (s Socket) Read(ch chan []byte, conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			// @todo Reconnect if this fails.
			log.Error().Msgf("Read error: %s", err)
		}
		ch <- message
	}
}

func (s Socket) Process(ch, conChan chan []byte) {
	go func() {
		for {
			payload := <-ch
			var txfr types.Txfr

			err := json.Unmarshal(payload, &txfr)
			if err != nil {
				log.Error().Msgf("There was an error while parsing the event: %v", err)
			}

			if txfr.Stream == "transfers" {
				if txfr.Data.FromAddr != ctypes.PoolAddress {
					var inTx ctypes.InTx

					for _, txn := range txfr.Data.T {
						// Temporary measure to get the memo.
						var tx types.Tx
						qp := map[string]string{}
						txDetails, _, _ := s.Binance.BasicClient.Get("/tx/"+txfr.Data.Hash+"?format=json", qp)
						json.Unmarshal(txDetails, &tx)

						txItem := ctypes.InTxItem{Tx: txfr.Data.Hash, 
							Memo: tx.Tx.Value.Memo,
							Sender: txfr.Data.FromAddr,
						}

						for _, coin := range txn.Coins {
							parsedAmt, _ := strconv.ParseFloat(coin.Amount, 64)
							amount := parsedAmt*100000000

							var token ctypes.Coins
							token.Denom = coin.Asset
							token.Amount = fmt.Sprintf("%.0f", amount)
							txItem.Coins = append(txItem.Coins, token)
						}

						inTx.TxArray = append(inTx.TxArray, txItem)
					}

					inTx.BlockHeight = txfr.Data.EventHeight
					inTx.Count = len(inTx.TxArray)

					json, err := json.Marshal(inTx)
					log.Info().Msgf("%v", string(json))
					if err != nil {
						log.Error().Msgf("Error: %v", err)
					}

					conChan <- json
				}
			}
		}
	}()
}

func (s Socket) Stop() {
	log.Info().Msgf("Shutting down....")
	os.Exit(1)
}
