package thorchain

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

// SetAdminConfig - saving a given admin config to the KVStore
func (k Keeper) SetAdminConfig(ctx sdk.Context, config AdminConfig) {
	store := ctx.KVStore(k.storeKey)
	key := getKey(prefixAdmin, config.DbKey(), getVersion(k.GetLowestActiveVersion(ctx), prefixAdmin))
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(config))
}

// GetAdminConfigValidatorRotateInNumBeforeFull
func (k Keeper) GetAdminConfigValidatorRotateInNumBeforeFull(ctx sdk.Context, addr sdk.AccAddress) int64 {
	return k.GetAdminConfigInt64(ctx, ValidatorRotateInNumBeforeFullKey, ValidatorRotateInNumBeforeFullKey.Default(), addr)
}

// GetAdminConfigValidatorRotateOutNumBeforeFull
func (k Keeper) GetAdminConfigValidatorRotateOutNumBeforeFull(ctx sdk.Context, addr sdk.AccAddress) int64 {
	return k.GetAdminConfigInt64(ctx, ValidatorRotateOutNumBeforeFullKey, ValidatorRotateOutNumBeforeFullKey.Default(), addr)
}

// GetAdminConfigValidatorRotateNumAfterFull
func (k Keeper) GetAdminConfigValidatorRotateNumAfterFull(ctx sdk.Context, addr sdk.AccAddress) int64 {
	return k.GetAdminConfigInt64(ctx, ValidatorRotateNumAfterFullKey, ValidatorRotateNumAfterFullKey.Default(), addr)
}

// GetAdminConfigDefaultPoolStatus - get the config for Default Pool Status
func (k Keeper) GetAdminConfigDefaultPoolStatus(ctx sdk.Context, addr sdk.AccAddress) PoolStatus {
	name, _ := k.GetAdminConfigValue(ctx, DefaultPoolStatus, addr)
	if name == "" {
		name = DefaultPoolStatus.Default()
	}
	return GetPoolStatus(name)
}

// GetAdminConfigGSL - get the config for GSL
func (k Keeper) GetAdminConfigGSL(ctx sdk.Context, addr sdk.AccAddress) common.Amount {
	return k.GetAdminConfigAmountType(ctx, GSLKey, GSLKey.Default(), addr)
}

// GetAdminConfigStakerAmtInterval - get the config for StakerAmtInterval
func (k Keeper) GetAdminConfigStakerAmtInterval(ctx sdk.Context, addr sdk.AccAddress) common.Amount {
	return k.GetAdminConfigAmountType(ctx, StakerAmtIntervalKey, StakerAmtIntervalKey.Default(), addr)
}

// GetAdminConfigMinValidatorBond get the minimum bond to become a validator
func (k Keeper) GetAdminConfigMinValidatorBond(ctx sdk.Context, addr sdk.AccAddress) sdk.Uint {
	return k.GetAdminConfigUintType(ctx, MinValidatorBondKey, MinValidatorBondKey.Default(), addr)
}

// GetAdminConfigWhiteListGasAsset
func (k Keeper) GetAdminConfigWhiteListGasAsset(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.GetAdminConfigCoinsType(ctx, WhiteListGasAssetKey, WhiteListGasAssetKey.Default(), addr)
}

// GetAdminConfigBnbAddressType - get the config with return type is BNBAddress
func (k Keeper) GetAdminConfigBnbAddressType(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) common.Address {
	value, _ := k.GetAdminConfigValue(ctx, key, addr)
	if value == "" {
		value = dValue
	}
	return common.Address(value)
}

// GetAdminConfigDesireValidatorSet
func (k Keeper) GetAdminConfigDesireValidatorSet(ctx sdk.Context, addr sdk.AccAddress) int64 {
	return k.GetAdminConfigInt64(ctx, DesireValidatorSetKey, DesireValidatorSetKey.Default(), addr)
}

// GetAdminConfigRotatePerBlockHeight get rotate per block height
func (k Keeper) GetAdminConfigRotatePerBlockHeight(ctx sdk.Context, addr sdk.AccAddress) int64 {
	return k.GetAdminConfigInt64(ctx, RotatePerBlockHeightKey, RotatePerBlockHeightKey.Default(), addr)
}

// GetAdminConfigValidatorsChangeWindow get validator change window
func (k Keeper) GetAdminConfigValidatorsChangeWindow(ctx sdk.Context, addr sdk.AccAddress) int64 {
	return k.GetAdminConfigInt64(ctx, ValidatorsChangeWindowKey, ValidatorsChangeWindowKey.Default(), addr)
}

func (k Keeper) GetAdminConfigUintType(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) sdk.Uint {
	value, _ := k.GetAdminConfigValue(ctx, key, addr)
	if value == "" {
		value = dValue
	}
	amt, err := common.NewAmount(value)
	if nil != err {
		ctx.Logger().Error("fail to parse value to float", "value", value)
	}
	return common.AmountToUint(amt)
}

// GetAdminConfigAmountType - get the config for TSL
func (k Keeper) GetAdminConfigAmountType(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) common.Amount {
	value, _ := k.GetAdminConfigValue(ctx, key, addr)
	if value == "" {
		value = dValue
	}
	return common.Amount(value)
}

// GetAdminConfigCoinsType - get the config for TSL
func (k Keeper) GetAdminConfigCoinsType(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) sdk.Coins {
	value, _ := k.GetAdminConfigValue(ctx, key, addr)
	if value == "" {
		value = dValue
	}
	coins, _ := sdk.ParseCoins(value)
	return coins
}

// GetAdminConfigInt64 - get the int64 config
func (k Keeper) GetAdminConfigInt64(ctx sdk.Context, key AdminConfigKey, dValue string, addr sdk.AccAddress) int64 {
	value, _ := k.GetAdminConfigValue(ctx, key, addr)
	if value == "" {
		value = dValue
	}
	result, _ := strconv.ParseInt(value, 10, 64)
	return result
}

// GetAdminConfigValue - gets the value of a given admin key
func (k Keeper) GetAdminConfigValue(ctx sdk.Context, kkey AdminConfigKey, addr sdk.AccAddress) (val string, err error) {
	getConfigValue := func(nodeAddr sdk.AccAddress) (string, error) {
		config := NewAdminConfig(kkey, "", nodeAddr)
		key := getKey(prefixAdmin, config.DbKey(), getVersion(k.GetLowestActiveVersion(ctx), prefixAdmin))
		store := ctx.KVStore(k.storeKey)
		if !store.Has([]byte(key)) {
			return kkey.Default(), nil
		}
		buf := store.Get([]byte(key))
		if err := k.cdc.UnmarshalBinaryBare(buf, &config); nil != err {
			ctx.Logger().Error(fmt.Sprintf("fail to unmarshal admin config, err: %s", err))
			return "", errors.Wrap(err, "fail to unmarshal admin config")
		}
		return config.Value, nil
	}
	// no specific bnb address given, look for consensus value
	if addr.Empty() {
		nodeAccounts, err := k.ListActiveNodeAccounts(ctx)
		if nil != err {
			return "", errors.Wrap(err, "fail to get active node accounts")
		}
		counter := make(map[string]int)
		for _, node := range nodeAccounts {
			config, err := getConfigValue(node.NodeAddress)
			if err != nil {
				return "", err
			}
			counter[config] += 1
		}

		for k, v := range counter {
			if HasMajority(v, len(nodeAccounts)) {
				return k, nil
			}
		}
	} else {
		// lookup admin config set by specific bnb address
		val, err = getConfigValue(addr)
		if err != nil {
			return val, err
		}
	}

	if val == "" {
		val = kkey.Default()
	}

	return val, err
}

// GetAdminConfigIterator iterate admin configs
func (k Keeper) GetAdminConfigIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixAdmin))
}
