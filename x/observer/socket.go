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

	types "gitlab.com/thorchain/bepswap/observe/x/observer/types"
)

type Socket struct {
	PoolAddress string
	DEXHost string
	PongWait time.Duration
}

func NewSocket(poolAddress, dexHost string) *Socket {
	return &Socket{
		PoolAddress: poolAddress,
		DEXHost: dexHost,
		PongWait: 30*time.Second,
	}
}

func (s *Socket) Start(conChan chan []byte) {
	conn, err := s.Connect()
	if err != nil {
		log.Fatal().Msgf("[OBSERVER] There was an error while starting: %v", err)
	}

	log.Info().Msgf("[OBSERVER] Setting a keepalive of %v", s.PongWait)
	s.SetKeepAlive(conn)

	ch := make(chan []byte)

	log.Info().Msg("[OBSERVER] Listening for events....")
	s.Process(ch, conChan)
	s.Read(ch, conn)
}

func (s *Socket) Connect() (*websocket.Conn, error) {
	path := fmt.Sprintf("/api/ws/%s", s.PoolAddress)
	url := url.URL{
		Scheme: "wss",
		Host: s.DEXHost,
		Path: path,
	}

	log.Info().Msgf("[OBSERVER] Opening up a connection to: %v", url.String())

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (s *Socket) SetKeepAlive(conn *websocket.Conn) {
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

func (s *Socket) Read(ch chan []byte, conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			// @todo Reconnect if this fails.
			log.Error().Msgf("[OBSERVER] Read error: %s", err)
		}
		ch <- message
	}
}

func (s *Socket) Process(ch, conChan chan []byte) {
	go func() {
		for {
			payload := <-ch
			var txfr types.Txfr

			err := json.Unmarshal(payload, &txfr)
			if err != nil {
				log.Error().Msgf("[OBSERVER] There was an error while parsing the event: %v", err)
			}

			if txfr.Stream == "transfers" {
				if txfr.Data.FromAddr != s.PoolAddress {
					var inTx types.InTx

					for _, txn := range txfr.Data.T {
						for _, coin := range txn.Coins {
							parsedAmt, _ := strconv.ParseFloat(coin.Amount, 64)
							amount := parsedAmt*100000000

							txItem := types.TxItem{Tx: txfr.Data.Hash, 
								Memo: "MEMO", //txfr.Data.Memo,
								Sender: txfr.Data.FromAddr,
								Coins: types.Coins{
									Denom: coin.Asset,
									Amount: fmt.Sprintf("%.0f", amount),
								},
							}

							inTx.TxArray = append(inTx.TxArray, txItem)
						}
					}

					inTx.BlockHeight = txfr.Data.EventHeight
					inTx.Count = len(inTx.TxArray)

					json, err := json.Marshal(inTx)
					if err != nil {
						log.Error().Msgf("[OBSERVER] Error: %v", err)
					}

					conChan <- json
				}
			}
		}
	}()
}

func (s *Socket) Stop() {
	log.Info().Msg("[OBSERVER] Shutting down....")
	os.Exit(1)
}
