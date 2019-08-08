package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const httpPort = 3737

// input Tx item
type txItem struct {
	To     string `json:"to"`
	Ticker string `json:"denom"`
	Amount string `json:"amount"`
}

// input sign request struct
type signRequest struct {
	Height  string   `json:"height"`
	Hash    string   `json:"hash"`
	TxArray []txItem `json:"tx_array"`
}

// response struct
type response struct {
	TxHash string `json:"tx_hash"`
}

// health check
func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"ping":"pong"}`)
}

// sign a request
func signHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	decoder := json.NewDecoder(r.Body)

	var t signRequest
	err := decoder.Decode(&t)

	if err != nil {
		panic(err)
	}

	// TODO: sign and broadcast the inputs

	resp := response{
		TxHash: "BOGUS_TX_HASH",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Logging requests...
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

// Error handler
func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(w, `{"code": 404, "error": "not found"}`)
	}
}

func main() {
	http.HandleFunc("/sign", signHandler)
	http.HandleFunc("/ping", pingHandler)

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	fmt.Printf("Listening on port %d...\n", httpPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), logRequest(http.DefaultServeMux))
	if err != nil {
		log.Fatal(err)
	}
}
