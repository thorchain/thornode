package rest

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gorilla/mux"
	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/query"
)

const (
	restURLParam  = "param1"
	restURLParam2 = "param2"
)

// TODO add stake record endpoint
// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {

	// limit api calls
	lmt := tollbooth.NewLimiter(10, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	lmt.SetMessage("You have reached maximum request limit.")

	// Health Check Endpoint
	r.Handle(
		fmt.Sprintf("/%s/ping", storeName),
		tollbooth.LimitFuncHandler(
			lmt,
			pingHandler(cliCtx, storeName),
		),
	).Methods("GET")

	// Dynamically create endpoints of all funcs in querier.go
	for _, q := range query.Queries {
		endpoint := q.Endpoint(storeName, restURLParam, restURLParam2)
		if endpoint != "" { // don't setup REST endpoint if we have no endpoint
			r.Handle(
				endpoint,
				tollbooth.LimitFuncHandler(
					lmt,
					getHandlerWrapper(q, storeName, cliCtx),
				),
			).Methods("GET")
		}
	}

	// Get unsigned json for emitting a binance transaction. Validators only.
	r.HandleFunc(
		fmt.Sprintf("/%s/binance/tx", storeName),
		postTxHashHandler(cliCtx),
	).Methods("POST")
}
