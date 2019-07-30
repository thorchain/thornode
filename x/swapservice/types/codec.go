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
	cdc.RegisterConcrete(PoolStruct{}, "swapservice/poolstruct", nil)
	cdc.RegisterConcrete(PoolStaker{}, "swapservice/poolstaker", nil)
	cdc.RegisterConcrete(StakerPool{}, "swapservice/stakerpool", nil)
	cdc.RegisterConcrete(PoolIndex{}, "swapservice/poolindex", nil)
	cdc.RegisterConcrete(StakerPoolItem{}, "swapservice/stakerpoolitem", nil)
	cdc.RegisterConcrete(StakerUnit{}, "swapservice/stakerunit", nil)
}
