package thorchain

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

type KeeperAdminConfig interface {
	SetAdminConfig(ctx sdk.Context, config AdminConfig)
	GetAdminConfigDefaultPoolStatus(ctx sdk.Context, addr sdk.AccAddress) PoolStatus
	GetAdminConfigValue(ctx sdk.Context, kkey AdminConfigKey, addr sdk.AccAddress) (val string, err error)
	GetAdminConfigIterator(ctx sdk.Context) sdk.Iterator
}

// SetAdminConfig - saving a given admin config to the KVStore
func (k KVStore) SetAdminConfig(ctx sdk.Context, config AdminConfig) {
	store := ctx.KVStore(k.storeKey)
	key := k.GetKey(ctx, prefixAdmin, config.DbKey())
	store.Set([]byte(key), k.cdc.MustMarshalBinaryBare(config))
}

// GetAdminConfigDefaultPoolStatus - get the config for Default Pool Status
func (k KVStore) GetAdminConfigDefaultPoolStatus(ctx sdk.Context, addr sdk.AccAddress) PoolStatus {
	name, _ := k.GetAdminConfigValue(ctx, DefaultPoolStatus, addr)
	if name == "" {
		name = DefaultPoolStatus.Default()
	}
	return GetPoolStatus(name)
}

// GetAdminConfigValue - gets the value of a given admin key
func (k KVStore) GetAdminConfigValue(ctx sdk.Context, kkey AdminConfigKey, addr sdk.AccAddress) (val string, err error) {
	getConfigValue := func(nodeAddr sdk.AccAddress) (string, error) {
		config := NewAdminConfig(kkey, "", nodeAddr)
		key := k.GetKey(ctx, prefixAdmin, config.DbKey())
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
func (k KVStore) GetAdminConfigIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(prefixAdmin))
}
