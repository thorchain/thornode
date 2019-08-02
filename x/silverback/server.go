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
		
		sChan := make(chan chan string)
		go s.PoolBal(sChan)

		http.HandleFunc("/pool", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			ws, _ := upgrader.Upgrade(w, r, nil)
			client := make(chan string, 1)
			sChan <- client

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

func (s *Server) PoolBal(sChan chan chan string) {
	var clients []chan string
	bChan := make(chan []byte, 1)

	go func (target chan []byte) {
			for {
				data := s.Pool.GetBal()
				b, _ := json.Marshal(data)

				time.Sleep(2 * time.Second)
				target <- b
			}
	}(bChan)

	for {
			select {
			case client, _ := <-sChan:
					clients = append(clients, client)
			case balances, _ := <-bChan:
					for _, c := range clients {
							c <- string(balances)
					}
			}
	}
}
