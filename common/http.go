package common

import (
	"fmt"
	"net/http"
	"encoding/json"

	log "github.com/rs/zerolog/log"

	types "gitlab.com/thorchain/bepswap/observe/common/types"
)

func StartWebServer() {
	http.HandleFunc("/", StatusHandler)

	err := http.ListenAndServe(":"+types.StatusPort, nil)
	if err != nil {
		log.Fatal().Msgf("Error: %v", err)
	}
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	var status types.Status
	status.State = "OK"

	json, _ := json.Marshal(status)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(json))
}
