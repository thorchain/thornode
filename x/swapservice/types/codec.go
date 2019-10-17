package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
}

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSetPoolData{}, "swapservice/SetPoolData", nil)
	cdc.RegisterConcrete(MsgSetStakeData{}, "swapservice/SetStakeData", nil)
	cdc.RegisterConcrete(MsgSwap{}, "swapservice/Swap", nil)
	cdc.RegisterConcrete(MsgSetUnStake{}, "swapservice/SetUnStake", nil)
	cdc.RegisterConcrete(MsgSetTxIn{}, "swapservice/MsgSetTxIn", nil)
	cdc.RegisterConcrete(MsgSetAdminConfig{}, "swapservice/MsgSetAdminConfig", nil)
	cdc.RegisterConcrete(MsgSetTrustAccount{}, "swapservice/MsgSetTrustAccount", nil)
	cdc.RegisterConcrete(MsgEndPool{}, "swapservice/MsgEndPool", nil)
}
