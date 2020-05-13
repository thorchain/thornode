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
	cdc.RegisterConcrete(MsgSetStakeData{}, "thorchain/SetStakeData", nil)
	cdc.RegisterConcrete(MsgSwap{}, "thorchain/Swap", nil)
	cdc.RegisterConcrete(MsgTssPool{}, "thorchain/TssPool", nil)
	cdc.RegisterConcrete(MsgTssKeysignFail{}, "thorchain/TssKeysignFail", nil)
	cdc.RegisterConcrete(MsgSetUnStake{}, "thorchain/SetUnStake", nil)
	cdc.RegisterConcrete(MsgObservedTxIn{}, "thorchain/ObservedTxIn", nil)
	cdc.RegisterConcrete(MsgObservedTxOut{}, "thorchain/ObservedTxOut", nil)
	cdc.RegisterConcrete(MsgSetNodeKeys{}, "thorchain/MsgSetNodeKeys", nil)
	cdc.RegisterConcrete(MsgAdd{}, "thorchain/MsgAdd", nil)
	cdc.RegisterConcrete(MsgBond{}, "thorchain/MsgBond", nil)
	cdc.RegisterConcrete(MsgLeave{}, "thorchain/MsgLeave", nil)
	cdc.RegisterConcrete(MsgNoOp{}, "thorchain/MsgNoOp", nil)
	cdc.RegisterConcrete(MsgOutboundTx{}, "thorchain/MsgOutboundTx", nil)
	cdc.RegisterConcrete(MsgSetVersion{}, "thorchain/MsgSetVersion", nil)
	cdc.RegisterConcrete(MsgSetIPAddress{}, "thorchain/MsgSetIPAddress", nil)
	cdc.RegisterConcrete(MsgYggdrasil{}, "thorchain/MsgYggdrasil", nil)
	cdc.RegisterConcrete(MsgReserveContributor{}, "thorchain/MsgReserveContributor", nil)
	cdc.RegisterConcrete(MsgErrataTx{}, "thorchain/MsgErrataTx", nil)
	cdc.RegisterConcrete(MsgBan{}, "thorchain/MsgBan", nil)
	cdc.RegisterConcrete(MsgSwitch{}, "thorchain/MsgSwitch", nil)
	cdc.RegisterConcrete(MsgMimir{}, "thorchain/MsgMimir", nil)
}
