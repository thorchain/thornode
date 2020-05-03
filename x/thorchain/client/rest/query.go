package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"gitlab.com/thorchain/thornode/x/thorchain/query"
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
		heightStr, ok := r.URL.Query()["height"]
		if ok && len(heightStr) > 0 {
			height, err := strconv.ParseInt(heightStr[0], 10, 64)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			cliCtx = cliCtx.WithHeight(height)
		}
		param := mux.Vars(r)[restURLParam]
		text, err := r.URL.MarshalBinary()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(q.Path(storeName, param, mux.Vars(r)[restURLParam2]), text)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(res)
	}
}
