package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"

	"gitlab.com/thorchain/thornode/x/thorchain"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
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

	path := fmt.Sprintf("./test/fixtures/endpoints/auth/accounts/template.json")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err.Error())
	}

	var d map[string]interface{}
	if err := json.Unmarshal(content, &d); err != nil {
		log.Println(err.Error())
	}

	// mod data with past in node_address
	result := d["result"].(map[string]interface{})
	value := result["value"].(map[string]interface{})
	value["address"] = node_address

	content, err = json.Marshal(d)
	if err != nil {
		log.Println(err.Error())
	}

	// spew.Dump(d)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(content))
}

func observerHandleFunc(w http.ResponseWriter, r *http.Request) {
	log.Println("observerHandleFunc HIT!")
	vars := mux.Vars(r)
	node_address := vars["node_address"]

	// path := fmt.Sprintf("./test/fixtures/endpoints/observer/%s.json", node_address)
	path := fmt.Sprintf("./test/fixtures/endpoints/observer/template.json")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintln(w, err.Error())
		return
	}

	var d types.NodeAccount
	if err := json.Unmarshal(content, &d); err != nil {
		log.Println(err.Error())
		fmt.Fprintln(w, err.Error())
		return
	}

	d.NodeAddress, err = sdk.AccAddressFromBech32(node_address)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintln(w, err.Error())
		return
	}

	content, err = json.Marshal(d)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintln(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(content))
}

func main() {
	thorchain.SetupConfigForTest()

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
