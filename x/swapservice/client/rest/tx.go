package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"gitlab.com/thorchain/bepswap/common"
	"gitlab.com/thorchain/bepswap/statechain/x/swapservice/types"
)

type txItem struct {
	TxHash string       `json:"tx"`
	Coins  common.Coins `json:"coins"`
	Memo   string       `json:"MEMO"`
	Sender string       `json:"sender"`
}

type txHashReq struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	Blockheight string       `json:"blockHeight"`
	Count       string       `json:"count"`
	TxArray     []txItem     `json:"txArray"`
}

func postTxHashHandler(cliCtx context.CLIContext) http.HandlerFunc {
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

		voters := make([]types.TxInVoter, len(req.TxArray))
		for i, tx := range req.TxArray {
			txID, err := common.NewTxID(tx.TxHash)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}

			bnbAddr, err := common.NewBnbAddress(tx.Sender)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}

			tx := types.NewTxIn(tx.Coins, tx.Memo, bnbAddr)

			voters[i] = types.NewTxInVoter(txID, []types.TxIn{tx})
		}

		// create the message
		msg := types.NewMsgSetTxIn(voters, addr)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}
