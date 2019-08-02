package silverback

import (
	"os"
	"time"
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
		var upgrader = websocket.Upgrader {
			ReadBufferSize: 1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r * http.Request) bool {
					return true
			},
		}
		
		svrChan := make(chan chan string)
		go s.PoolBal(svrChan)

		http.HandleFunc("/pool", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			ws, _ := upgrader.Upgrade(w, r, nil)
			client := make(chan string, 1)
			svrChan <- client

			for {
					select {
					case text, _ := <-client:
						writer, _ := ws.NextWriter(websocket.TextMessage)
						writer.Write([]byte(text))
						writer.Close()
					}
			}
		})
		http.ListenAndServe(":" + s.Port, nil)
	}()
}

func (s *Server) PoolBal(svrChan chan chan string) {
	var clients []chan string
	balChan := make(chan []byte, 1)

	go func (target chan []byte) {
			for {
				data := s.Pool.GetBal()
				log.Info().Msgf("Broadcasting balances: %v", data)
				b, _ := json.Marshal(data)

				time.Sleep(5 * time.Second)
				target <- b
			}
	}(balChan)

	for {
			select {
			case client, _ := <-svrChan:
					clients = append(clients, client)
			case balances, _ := <-balChan:
					for _, c := range clients {
							c <- string(balances)
					}
			}
	}
}
