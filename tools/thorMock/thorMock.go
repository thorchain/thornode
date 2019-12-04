package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func poolAddressesMocked(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit poolAddressesMocked!!")
	content, err := ioutil.ReadFile("./test/fixtures/endpoints/poolAddresses/pooladdresses.json")
	if err != nil {
		log.Fatal(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(content))
}

func lastblockMocked(w http.ResponseWriter, r *http.Request) {
	log.Println("lastblockMocked HIT!")
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

func main() {
	addr := ":1317"
	router := mux.NewRouter()
	router.HandleFunc("/thorchain/pooladdresses", poolAddressesMocked).Methods("GET")
	router.HandleFunc("/thorchain/lastblock/{chain}", lastblockMocked).Methods("GET")

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	fmt.Println("Running thorMocked: ", addr)
	log.Fatal(srv.ListenAndServe())
}
