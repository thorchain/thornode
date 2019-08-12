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
	restTx         = "tx"
	restPoolStruct = "poolstruct"
	restPoolData   = "pooldata"
	restPoolStaker = "poolstaker"
	restStakerPool = "stakerpool"
	swapData       = "swapdata"
	stakeData      = "stakedata"
	accData        = "accdata"
	txoutArrayData = "txoutarray"
	adminConfig    = "adminconfig"
)

// pool index etc
// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	r.HandleFunc(fmt.Sprintf("/%s/admin/{%s}", storeName, adminConfig), getAdminConfig(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/tx/{%s}", storeName, restTx), getTxHash(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/pool/{%s}", storeName, restPoolStruct), poolStructHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/pool/{%s}/stakers", storeName, restPoolStaker), poolStakersHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/staker/{%s}", storeName, restStakerPool), stakerPoolHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/pools", storeName), poolHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/account/{%s}", storeName, accData), accHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/stake/{%s}", storeName, stakeData), stakeHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/swaprecord/{%s}", storeName, swapData), swapRecordHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/unstakerecord/{%s}", storeName, swapData), unStakeRecordHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/stake", storeName), setStakeDataHandler(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/%s/binance/tx", storeName), txHashHandler(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/%s/txoutarray/{%s}", storeName, txoutArrayData), txOutArrayHandler(cliCtx, storeName)).Methods("GET")
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

type setStakeData struct {
	BaseReq       rest.BaseReq `json:"base_req"`
	Ticker        string       `json:"ticker"`
	Rune          string       `json:"rune_amount"`
	Token         string       `json:"token_amount"`
	PublicAddress string       `json:"public_address"`
	RequestTxHash string       `json:"request_tx_hash"`
}

func setStakeDataHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		var req setStakeData
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		ticker, err := types.NewTicker(req.Ticker)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		txID, err := types.NewTxID(req.RequestTxHash)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		runeAmount, err := types.NewAmount(req.Rune)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tokenAmount, err := types.NewAmount(req.Token)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		bnbAddr, err := types.NewBnbAddress(req.PublicAddress)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := types.NewMsgSetStakeData(ticker, runeAmount, tokenAmount, bnbAddr, txID, cliCtx.GetFromAddress())
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

func getTxHash(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restTx]
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
		paramType := vars[restPoolStruct]
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
		paramType := vars[restPoolStaker]
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
		paramType := vars[restPoolData]
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

func accHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[accData]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/accstruct/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func stakeHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[stakeData]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/stakestruct/%s", storeName, paramType), nil)
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
		paramType := vars[swapData]
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
		paramType := vars[swapData]
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
		paramType := vars[txoutArrayData]
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
		paramType := vars[adminConfig]
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/adminconfig/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
