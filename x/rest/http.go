package rest

import (
	"net/http"
	"encoding/json"

	log "github.com/rs/zerolog/log"

	types "gitlab.com/thorchain/bepswap/observe/x/rest/types"
)

func Start(port string) {
	log.Info().Msg("Starting Rest....")
	go func() {
  	http.HandleFunc("/", Status)
		http.ListenAndServe(":" + port, nil)
	}()
}

func Status(w http.ResponseWriter, r *http.Request) {
	data := types.Response{Status: "OK"}

  js, err := json.Marshal(data)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  w.Write(js)
}
