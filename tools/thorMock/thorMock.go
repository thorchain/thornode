package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func poolAddressesHandleFunc(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit poolAddressesHandleFunc!!")
	content, err := ioutil.ReadFile("./test/fixtures/endpoints/poolAddresses/pooladdresses.json")
	if err != nil {
		log.Fatal(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(content))
}

func lastblockHandleFunc(w http.ResponseWriter, r *http.Request) {
	log.Println("lastblockHandleFunc HIT!")
	vars := mux.Vars(r)

	chain := vars["chain"]

	path := fmt.Sprintf("./test/fixtures/endpoints/lastblock/%v.json", strings.ToLower(chain))
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(content))
}

func vaultsPubKeysHandleFunc(w http.ResponseWriter, r *http.Request) {
	log.Println("lastblockHandleFunc HIT!")
	path := fmt.Sprintf("./test/fixtures/endpoints/vaults/pubKeys.json")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(content))
}

func authAccountsHandleFunc(w http.ResponseWriter, r *http.Request) {
	log.Println("authAccountsHandleFunc HIT!")
	vars := mux.Vars(r)
	node_address := vars["node_address"]

	path := fmt.Sprintf("./test/fixtures/endpoints/auth/accounts/%s.json", node_address)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(content))
}

func observerHandleFunc(w http.ResponseWriter, r *http.Request) {
	log.Println("observerHandleFunc HIT!")
	vars := mux.Vars(r)
	node_address := vars["node_address"]

	path := fmt.Sprintf("./test/fixtures/endpoints/observer/%s.json", node_address)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(content))
}

func main() {
	addr := ":1317"
	router := mux.NewRouter()
	router.HandleFunc("/thorchain/pool_addresses", poolAddressesHandleFunc).Methods("GET")
	router.HandleFunc("/thorchain/lastblock/{chain}", lastblockHandleFunc).Methods("GET")
	router.HandleFunc("/thorchain/vaults/pubkeys", vaultsPubKeysHandleFunc).Methods("GET")
	router.HandleFunc("/auth/accounts/{node_address}", authAccountsHandleFunc).Methods("GET")
	router.HandleFunc("/thorchain/observer/{node_address}", observerHandleFunc).Methods("GET")

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	fmt.Println("Running thorMocked: ", addr)
	log.Fatal(srv.ListenAndServe())
}
