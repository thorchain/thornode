package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"gitlab.com/thorchain/tss/go-tss/blame"

	"gitlab.com/thorchain/thornode/common"
	"gitlab.com/thorchain/thornode/x/thorchain/types"
)

type nativeTx struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Coins   common.Coins `json:"coins"`
}

func newNativeTxHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req nativeTx

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}
		baseReq.Gas = "auto"
		addr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := types.NewMsgNativeTx(req.Coins, req.BaseReq.Memo, addr)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

type newErrataTx struct {
	BaseReq rest.BaseReq `json:"base_req"`
	TxID    common.TxID  `json:"txid"`
	Chain   common.Chain `json:"chain"`
}

func newErrataTxHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req newErrataTx

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}
		baseReq.Gas = "auto"
		addr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		msg := types.NewMsgErrataTx(req.TxID, req.Chain, addr)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

type newTssPool struct {
	BaseReq      rest.BaseReq     `json:"base_req"`
	InputPubKeys common.PubKeys   `json:"input_pubkeys"`
	KeygenType   types.KeygenType `json:"keygen_type"`
	Height       int64            `json:"height"`
	Blame        blame.Blame      `json:"blame"`
	PoolPubKey   common.PubKey    `json:"pool_pub_key"`
	Chains       common.Chains    `json:"chains"`
}

func newTssPoolHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req newTssPool

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}
		baseReq.Gas = "auto"
		addr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		msg := types.NewMsgTssPool(req.InputPubKeys, req.PoolPubKey, req.KeygenType, req.Height, req.Blame, req.Chains, addr)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

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
		baseReq.Gas = "auto"
		addr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var inbound types.ObservedTxs
		var outbound types.ObservedTxs

		for _, tx := range req.Txs {
			chain := common.EmptyChain
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
				rest.WriteErrorResponse(w, http.StatusBadRequest, "Unable to determine the direction of observation")
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
