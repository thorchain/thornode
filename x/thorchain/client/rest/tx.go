package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

type txHashReq struct {
	BaseReq rest.BaseReq      `json:"base_req"`
	Txs     types.ObservedTxs `json:"txs"`
}

func postTxsHandler(cliCtx context.CLIContext) http.HandlerFunc {
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

		baseReq.Gas = "400000" // i think we can delete this "auto" gas should work

		addr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var inbound types.ObservedTxs
		var outbound types.ObservedTxs

		for _, tx := range req.Txs {
			chain := common.BNBChain
			if len(tx.Tx.Coins) > 0 {
				chain = tx.Tx.Coins[0].Asset.Chain
			}

			obAddr, err := tx.ObservedPubKey.GetAddress(chain)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			if tx.Tx.ToAddress.Equals(obAddr) {
				inbound = append(inbound, tx)
			} else if tx.Tx.FromAddress.Equals(obAddr) {
				outbound = append(outbound, tx)
			} else {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		msgs := make([]sdk.Msg, 0)

		if len(inbound) > 0 {
			msg := types.NewMsgObservedTxIn(inbound, addr)
			err = msg.ValidateBasic()
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			msgs = append(msgs, msg)
		}

		if len(outbound) > 0 {
			msg := types.NewMsgObservedTxOut(outbound, addr)
			err = msg.ValidateBasic()
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			msgs = append(msgs, msg)
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, msgs)
	}
}
