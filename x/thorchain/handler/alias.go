package handler

import (
	"gitlab.com/thorchain/bepswap/thornode/x/thorchain"
	"gitlab.com/thorchain/bepswap/thornode/x/thorchain/types"
)

const (
	DefaultCodespace = types.DefaultCodespace

	NodeActive = types.Active
)

type (
	// Standard Types
	Keeper             = thorchain.Keeper
	PoolAddressManager = thorchain.PoolAddressManager
	TxOutStore         = thorchain.TxOutStore
	ValidatorManager   = thorchain.ValidatorManager

	// Message Types
	MsgSetPoolData = types.MsgSetPoolData
)
