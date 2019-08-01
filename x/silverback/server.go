package silverback

import (
	"os"
	"net/http"
	"encoding/json"

	"github.com/gorilla/websocket"
	log "github.com/rs/zerolog/log"
)

type Server struct {
	Binance Binance
	Pool Pool
	Port string
}

func NewServer(binance Binance, pool Pool) *Server {
	return &Server{
		Binance: binance,
		Pool: pool,
		Port: os.Getenv("PORT"),
	}
}

func (s *Server) Start() {
	go func() {
		log.Info().Msg("Starting Silverback Server....")
		http.HandleFunc("/", s.Balances)
		http.ListenAndServe(":" + s.Port, nil)
	}()
}

func (s *Server) Balances(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Info().Msgf("Upgrade: %v", err)
		return
	}

	defer c.Close()
	for {
		mt, _, err := c.ReadMessage()
		if err != nil {
			log.Error().Msgf("Read error: %v", err)
			break
		}

		log.Info().Msgf("Received message: %v", mt)

		pool := NewPool(s.Binance.PoolAddress)
		data := pool.GetBal()

		js, err := json.Marshal(data)
		if err != nil {
			log.Error().Msgf("Marshalling error: %v", err)
			return
		}

		err = c.WriteMessage(mt, js)
		if err != nil {
			log.Error().Msgf("Write error: %v", err)
			break
		}
	}
}
