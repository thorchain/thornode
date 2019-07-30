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

func Start(pongWait time.Duration, poolAddress string, dexHost string) {
	log.Info().Msg("Starting Silverback....")
	
	conn, err := Connect(pongWait, poolAddress, dexHost)
	if err != nil {
		log.Error().Msgf("There was an error while starting: %v", err)
	}

	log.Info().Msgf("Setting a keepalive of %v", pongWait)
	SetKeepAlive(conn, pongWait)

	ch := make(chan []byte)

	log.Info().Msg("Listening for events....")
	ParseEvents(ch)
	ReadEvents(ch, conn)
}

func Connect(pongWait time.Duration, poolAddress string, dexHost string) (*websocket.Conn, error) {
	path := fmt.Sprintf("/api/ws/%s", poolAddress)
	url := url.URL{ Scheme: "wss", Host: dexHost, Path: path }

	log.Info().Msgf("Opening up a connection to: %v", url.String())

	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func SetKeepAlive(conn *websocket.Conn, pongWait time.Duration) {
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

			time.Sleep(pongWait / 2)
			if time.Now().Sub(lastResponse) > pongWait {
				conn.Close()
				return
			}
		}
	}()
}

func ReadEvents(ch chan []byte, conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Error().Msgf("Read error: %s", err)
		}
		ch <- message
	}
}

func ParseEvents(ch chan []byte) {
	go func() {
		for {
			payload := <-ch
			log.Info().Msgf("An event was received: %v", payload)

			var acct types.Account

			err := json.Unmarshal(payload, &acct)
			if err != nil {
				log.Error().Msgf("There was an while parsing the event: %v", err)
			}
		}
	}()
}

func Stop() {
	log.Info().Msg("Shutting down....")
	os.Exit(1)
}
