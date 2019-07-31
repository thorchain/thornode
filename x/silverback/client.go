package silverback

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/rs/zerolog/log"

	types "gitlab.com/thorchain/bepswap/observe/x/silverback/types"
)

type Client struct {
	PongWait time.Duration
	Binance Binance
}

func NewClient(binance Binance) *Client {
	pongWait := 30 * time.Second
	return &Client{
		Binance: binance,
		PongWait: pongWait,
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
			log.Info().Msgf("An event was received: %v", payload)

			var acct types.Account

			err := json.Unmarshal(payload, &acct)
			if err != nil {
				log.Error().Msgf("There was an error while parsing the event: %v", err)
			}
		}
	}()
}

func (c *Client) Stop() {
	log.Info().Msg("Shutting down....")
	os.Exit(1)
}
