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

func NewSocket(poolAddress, dexHost string) *Socket {
	binance := binance.NewBinance(poolAddress, dexHost)
	return &Socket{
		Binance: binance,
		PongWait: 30*time.Second,
	}
}

func (s *Socket) Start(conChan chan []byte) {
	conn, err := s.Connect()
	if err != nil {
		log.Fatal().Msgf("%s There was an error while starting: %v", LogPrefix(), err)
	}

	log.Info().Msgf("%s Setting a keepalive of %v", LogPrefix(), s.PongWait)
	s.SetKeepAlive(conn)

	ch := make(chan []byte)

	log.Info().Msgf("%s Listening for events....", LogPrefix())
	s.Process(ch, conChan)
	s.Read(ch, conn)
}

func (s *Socket) Connect() (*websocket.Conn, error) {
	path := fmt.Sprintf("/api/ws/%s", s.Binance.PoolAddress)
	url := url.URL{
		Scheme: "wss",
		Host: s.Binance.DEXHost,
		Path: path,
	}

	log.Info().Msgf("%s Opening up a connection to: %v", LogPrefix(), url.String())

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
			log.Error().Msgf("%s Read error: %s", LogPrefix(), err)
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
				log.Error().Msgf("%s There was an error while parsing the event: %v", LogPrefix(), err)
			}

			if txfr.Stream == "transfers" {
				if txfr.Data.FromAddr != s.Binance.PoolAddress {
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
						log.Error().Msgf("%s Error: %v", LogPrefix(), err)
					}

					conChan <- json
				}
			}
		}
	}()
}

func (s *Socket) Stop() {
	log.Info().Msgf("%s Shutting down....", LogPrefix())
	os.Exit(1)
}
