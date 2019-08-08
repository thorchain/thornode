package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/gorilla/mux"

	"gitlab.com/thorchain/statechain/x/swapservice/types"
)

const (
	restURLParam = "param1"
)

// TODO add stake record endpoint
// TODO add pool index endpoint
// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	r.HandleFunc(fmt.Sprintf("/%s/ping", storeName), pingHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/admin/{%s}", storeName, restURLParam), getAdminConfig(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/tx/{%s}", storeName, restURLParam), getTxHash(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/pool/{%s}", storeName, restURLParam), poolStructHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/pool/{%s}/stakers", storeName, restURLParam), poolStakersHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/staker/{%s}", storeName, restURLParam), stakerPoolHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/pools", storeName), poolHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/swaprecord/{%s}", storeName, restURLParam), swapRecordHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/unstakerecord/{%s}", storeName, restURLParam), unStakeRecordHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/binance/tx", storeName), txHashHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/txoutarray/{%s}", storeName, restURLParam), txOutArrayHandler(cliCtx, storeName)).Methods("GET")
}

type txItem struct {
	TxHash string      `json:"tx"`
	Coins  types.Coins `json:"coins"`
	Memo   string      `json:"MEMO"`
	Sender string      `json:"sender"`
}

type txHashReq struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	Blockheight int          `json:"blockHeight"`
	Count       int          `json:"count"`
	TxArray     []txItem     `json:"txArray"`
}

func txHashHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req txHashReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		addr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		txHashes := make([]types.TxHash, len(req.TxArray))
		for i, tx := range req.TxArray {
			txID, err := types.NewTxID(tx.TxHash)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}

			bnbAddr, err := types.NewBnbAddress(tx.Sender)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}

			txHashes[i] = types.NewTxHash(txID, tx.Coins, tx.Memo, bnbAddr)
		}

		// create the message
		msg := types.NewMsgSetTxHash(txHashes, addr)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

// Ping - endpoint to check that the API is up and available
func pingHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(`{"ping":"pong"}`))
	}
}

func getTxHash(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restURLParam]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/txhash/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func poolStructHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restURLParam]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/poolstruct/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func poolStakersHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restURLParam]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/poolstakers/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func stakerPoolHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restURLParam]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/stakerpools/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
func poolHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/pools", storeName), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func swapRecordHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restURLParam]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/swaprecord/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func unStakeRecordHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restURLParam]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/unstakerecord/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func txOutArrayHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restURLParam]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/txoutarray/%s", storeName, paramType), nil)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getAdminConfig(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restURLParam]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/adminconfig/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
