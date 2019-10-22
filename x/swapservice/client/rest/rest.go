package rest

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gorilla/mux"
	"gitlab.com/thorchain/bepswap/thornode/x/swapservice/query"
)

const (
	restURLParam  = "param1"
	restURLParam2 = "param2"
)

// TODO add stake record endpoint
// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {

	// Health Check Endpoint
	r.HandleFunc(
		fmt.Sprintf("/%s/ping", storeName),
		pingHandler(cliCtx, storeName),
	).Methods("GET")

	// limit api calls
	lmt := tollbooth.NewLimiter(1, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	lmt.SetMessage("You have reached maximum request limit.")

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
