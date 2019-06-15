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
	cdc.RegisterConcrete(MsgBuyPoolData{}, "swapservice/BuyPoolData", nil)
}
