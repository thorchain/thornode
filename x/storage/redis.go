package storage

import (
	"github.com/go-redis/redis"

	types "gitlab.com/thorchain/bepswap/observe/common/types"
)

func RedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     types.RedisUrl,
		Password: types.RedisPasswd,
		DB:       0,
	})

	return client
}
