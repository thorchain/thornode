package rest

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/gorilla/mux"
	"gitlab.com/thorchain/statechain/x/swapservice/query"
)

const (
	restURLParam = "param1"
)

// TODO add stake record endpoint
// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {

	// Health Check Endpoint
	r.HandleFunc(
		fmt.Sprintf("/%s/ping", storeName),
		pingHandler(cliCtx, storeName),
	).Methods("GET")

	// Dynamically create endpoints of all funcs in querier.go
	for _, q := range query.Queries {
		r.HandleFunc(
			q.Sprintf(storeName, restURLParam),
			getHandlerWrapper(q, storeName, cliCtx),
		).Methods("GET")
	}

	// Get unsigned json for emitting a binance transaction. Validators only.
	r.HandleFunc(
		fmt.Sprintf("/%s/binance/tx", storeName),
		postTxHashHandler(cliCtx),
	).Methods("POST")
}
