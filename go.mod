module gitlab.com/thorchain/bepswap/observe

go 1.12

require (
	bou.ke/monkey v1.0.1 // indirect
	github.com/avast/retry-go v2.4.1+incompatible
	github.com/binance-chain/go-sdk v1.0.8
	github.com/cosmos/cosmos-sdk v0.36.0-rc1
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/gorilla/websocket v1.4.0
	github.com/hashicorp/go-retryablehttp v0.5.4
	github.com/matryer/try v0.0.0-20161228173917-9ac251b645a2
	github.com/otiai10/copy v0.0.0-20180813032824-7e9a647135a1 // indirect
	github.com/otiai10/curr v0.0.0-20150429015615-9b4961190c95 // indirect
	github.com/otiai10/mint v1.2.3 // indirect
	github.com/pkg/errors v0.8.1
	github.com/r3labs/diff v0.0.0-20190801153147-a71de73c46ad
	github.com/rs/zerolog v1.14.3
	github.com/syndtr/goleveldb v1.0.1-0.20190318030020-c3a204f8e965
	github.com/valyala/fasthttp v1.4.0
	gitlab.com/thorchain/bepswap/common v0.0.0-20190816093251-b84c21cee45c
	gitlab.com/thorchain/bepswap/statechain v0.0.0-00010101000000-000000000000
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
)

replace gitlab.com/thorchain/bepswap/observe => ../observe

replace gitlab.com/thorchain/bepswap/statechain => ../statechain

replace github.com/tendermint/go-amino => github.com/binance-chain/bnc-go-amino v0.14.1-binance.1
