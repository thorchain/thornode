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
	cdc.RegisterConcrete(MsgSetPoolData{}, "thorchain/SetPoolData", nil)
	cdc.RegisterConcrete(MsgSetStakeData{}, "thorchain/SetStakeData", nil)
	cdc.RegisterConcrete(MsgSwap{}, "thorchain/Swap", nil)
	cdc.RegisterConcrete(MsgSetUnStake{}, "thorchain/SetUnStake", nil)
	cdc.RegisterConcrete(MsgObservedTxIn{}, "thorchain/ObservedTxIn", nil)
	cdc.RegisterConcrete(MsgObservedTxOut{}, "thorchain/ObservedTxOut", nil)
	cdc.RegisterConcrete(MsgSetAdminConfig{}, "thorchain/MsgSetAdminConfig", nil)
	cdc.RegisterConcrete(MsgSetTrustAccount{}, "thorchain/MsgSetTrustAccount", nil)
	cdc.RegisterConcrete(MsgEndPool{}, "thorchain/MsgEndPool", nil)
	cdc.RegisterConcrete(MsgAdd{}, "thorchain/MsgAdd", nil)
	cdc.RegisterConcrete(MsgBond{}, "thorchain/MsgBond", nil)
	cdc.RegisterConcrete(MsgLeave{}, "thorchain/MsgLeave", nil)
	cdc.RegisterConcrete(MsgNoOp{}, "thorchain/MsgNoOp", nil)
	cdc.RegisterConcrete(MsgOutboundTx{}, "thorchain/MsgOutboundTx", nil)
	cdc.RegisterConcrete(MsgSetVersion{}, "thorchain/MsgSetVersion", nil)
	cdc.RegisterConcrete(MsgYggdrasil{}, "thorchain/MsgYggdrasil", nil)
	cdc.RegisterConcrete(MsgReserveContributor{}, "thorchain/MsgReserveContributor", nil)
}
