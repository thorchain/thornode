package silverback

import (
	"net/http"
	"encoding/json"
	"math/rand"

	"github.com/gorilla/websocket"
	log "github.com/rs/zerolog/log"

	types "gitlab.com/thorchain/bepswap/observe/x/silverback/types"
)

type server struct {
	port string
}

func NewServer(port string) *server {
	return &server{
		port: port,
	}
}

func (s *server) Start() {
	go func() {
		log.Info().Msg("Starting Silverback Server....")
  	http.HandleFunc("/", s.Calc)
		http.ListenAndServe(":" + s.port, nil)
	}()
}

func (s *server) Calc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	upgrader := websocket.Upgrader{}

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

		log.Info().Msgf("Received: %v", err)

		data := types.Calc{
			X: rand.Float64(),
			Y: rand.Float64(),
			R: rand.Float64(),
			Z: rand.Float64(),
		}

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
