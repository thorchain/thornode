module gitlab.com/thorchain/bepswap/observe

go 1.12

require (
	github.com/avast/retry-go v2.4.1+incompatible
	github.com/binance-chain/go-sdk v1.0.8
	github.com/cosmos/cosmos-sdk v0.35.0 // indirect
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/gorilla/websocket v1.4.0
	github.com/hashicorp/go-retryablehttp v0.5.4
	github.com/matryer/try v0.0.0-20161228173917-9ac251b645a2
	github.com/r3labs/diff v0.0.0-20190801153147-a71de73c46ad
	github.com/rs/zerolog v1.14.3
)

replace gitlab.com/thorchain/bepswap/observe => ../observe

replace github.com/tendermint/go-amino => github.com/binance-chain/bnc-go-amino v0.14.1-binance.1
