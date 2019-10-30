package rest

import (
	"errors"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"gitlab.com/thorchain/bepswap/thornode/common"

	"gitlab.com/thorchain/bepswap/thornode/x/swapservice/types"
)

type txItem struct {
	TxHash             string       `json:"tx"`
	Coins              common.Coins `json:"coins"`
	Memo               string       `json:"MEMO"`
	Sender             string       `json:"sender"`
	ObservePoolAddress string       `json:"observe_pool_address"`
}

type txHashReq struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	Blockheight string       `json:"blockHeight"`
	Count       string       `json:"count"`
	Chain       string       `json:"chain"`
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

		voters := make([]types.TxInVoter, 0)
		height := sdk.NewUintFromString(req.Blockheight)
		if height.IsZero() {
			err := errors.New("chain block height cannot be zero")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		for _, tx := range req.TxArray {
			txID, err := common.NewTxID(tx.TxHash)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}

			bnbAddr, err := common.NewAddress(tx.Sender)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			observeAddr, err := common.NewPubKeyFromHexString(tx.ObservePoolAddress)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			}

			tx := types.NewTxIn(tx.Coins, tx.Memo, bnbAddr, height, observeAddr)

			voters = append(voters, types.NewTxInVoter(txID, []types.TxIn{tx}))
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
