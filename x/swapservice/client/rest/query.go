package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"gitlab.com/thorchain/bepswap/thor-node/x/swapservice/query"
)

// Ping - endpoint to check that the API is up and available
func pingHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"ping":"pong"}`)
	}
}

// Generic wrapper to generate GET handler
func getHandlerWrapper(q query.Query, storeName string, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		param := mux.Vars(r)[restURLParam]
		res, _, err := cliCtx.QueryWithData(q.Path(storeName, param), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(res)

	}
}
