package silverback

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"
	"strconv"

	"github.com/gorilla/websocket"
	log "github.com/rs/zerolog/log"

	types "gitlab.com/thorchain/bepswap/observe/x/silverback/types"
)

type Client struct {
	Binance Binance
	Pool Pool
	PongWait time.Duration
}

func NewClient(binance Binance, pool Pool) *Client {
	return &Client{
		Binance: binance,
		Pool: pool,
		PongWait: 30 * time.Second,
	}
}

func (c *Client) Start() {
	log.Info().Msg("Starting Silverback Client....")

	conn, err := c.Connect()
	if err != nil {
		log.Error().Msgf("There was an error while starting: %v", err)
	}

	log.Info().Msgf("Setting a keepalive of %v", c.PongWait)
	c.SetKeepAlive(conn)

	ch := make(chan []byte)

	log.Info().Msg("Listening for events....")
	c.ParseEvents(ch)
	c.ReadEvents(ch, conn)
}

func (c *Client) Connect() (*websocket.Conn, error) {
	path := fmt.Sprintf("/api/ws/%s", c.Binance.PoolAddress)
	url := url.URL{ Scheme: "wss", Host: c.Binance.DexHost, Path: path }

	log.Info().Msgf("Opening up a connection to: %v", url.String())

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Client) SetKeepAlive(conn *websocket.Conn) {
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

			time.Sleep(c.PongWait / 2)
			if time.Now().Sub(lastResponse) > c.PongWait {
				conn.Close()
				return
			}
		}
	}()
}

func (c *Client) ReadEvents(ch chan []byte, conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Error().Msgf("Read error: %s", err)
		}
		ch <- message
	}
}

func (c *Client) ParseEvents(ch chan []byte) {
	go func() {
		for {
			payload := <-ch
			var resp types.Response

			err := json.Unmarshal(payload, &resp)
			if err != nil {
				log.Error().Msgf("There was an error while parsing the event: %v", err)
			}

			switch resp.Stream {
			case "accounts":
				var acct types.Account

				json.Unmarshal(payload, &acct)
				if acct.Data.EventType == "outboundAccountInfo" {
					SyncBal(c.Binance)
				}
			case "transfers":
				var tnsfr types.Transfer
				json.Unmarshal(payload, &tnsfr)

				if tnsfr.Data.FromAddr != c.Binance.PoolAddress {
					log.Info().Msgf("Event received: %s", string(payload))

					if tnsfr.Data.EventType == "outboundTransferInfo" {
						fromAddr := tnsfr.Data.FromAddr
						for _, tx := range tnsfr.Data.T {
							for _, coin := range tx.Coins {
								c.ProcessTxn(fromAddr, coin.Asset, coin.Amount)
							}
						}
					}
				}
			}
		}
	}()
}

func (c *Client) ProcessTxn(fromAddress, symbol, amount string) {
	x, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
	}

	X, Y, txnAsset := c.CalcVars(symbol, amount)

	log.Info().Msgf("CalcOutput: %v", c.Pool.CalcOutput(x, X, Y))
	log.Info().Msgf("CalcOutputSlip: %v", c.Pool.CalcOutputSlip(x, X))
	log.Info().Msgf("CalcLiquidityFee: %v", c.Pool.CalcLiquidityFee(x, X, Y))

	emitTokens := c.Pool.CalcTokensEmitted(x, X, Y)
	log.Info().Msgf("CalcTokensEmitted: %v", emitTokens)
	log.Info().Msgf("CalcTradeSlip: %v", c.Pool.CalcTradeSlip(x, X, Y))
	log.Info().Msgf("CalcPoolSlip: %v", c.Pool.CalcPoolSlip(x, X, Y))

	c.Binance.SendToken(fromAddress, txnAsset, int64(emitTokens * 100000000))
}

func (c *Client) CalcVars(symbol, amount string) (float64, float64, string) {
	var (
		X float64
		Y float64
		txnAsset string
	)

	balances := c.Pool.GetBal()

	if symbol == c.Pool.X {
		X, _ = strconv.ParseFloat(balances.X, 64)
		Y, _ = strconv.ParseFloat(balances.Y, 64)
		txnAsset = c.Pool.Y
	} else if symbol == c.Pool.Y {
		X, _ = strconv.ParseFloat(balances.Y, 64)
		Y, _ = strconv.ParseFloat(balances.X, 64)
		txnAsset = c.Pool.X
	}

	return X, Y, txnAsset
}

func (c *Client) Stop() {
	log.Info().Msg("Shutting down....")
	os.Exit(1)
}
