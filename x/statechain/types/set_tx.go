package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type SetTx struct {
	Mode string `json:"mode"`
	Tx   struct {
		Msg        []sdk.Msg                `json:"msg"`
		Fee        authtypes.StdFee         `json:"fee"`
		Signatures []authtypes.StdSignature `json:"signatures"`
		Memo       string                   `json:"memo"`
	} `json:"tx"`
}
