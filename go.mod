module gitlab.com/thorchain/bepswap/observe

go 1.12

require (
	github.com/binance-chain/go-sdk v1.0.8
	github.com/cosmos/cosmos-sdk v0.36.0-rc1
	github.com/gorilla/websocket v1.4.0
	github.com/hashicorp/go-retryablehttp v0.5.4
	github.com/jinzhu/configor v1.1.1
	github.com/pkg/errors v0.8.1
	github.com/rs/zerolog v1.14.3
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	github.com/syndtr/goleveldb v1.0.1-0.20190318030020-c3a204f8e965
	github.com/valyala/fasthttp v1.4.0
	gitlab.com/thorchain/bepswap/common v0.0.0-20190816093251-b84c21cee45c
	gitlab.com/thorchain/bepswap/statechain v0.0.0-20190817014219-e1e1a77d6935
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
)

replace gitlab.com/thorchain/bepswap/observe => ../observe

replace github.com/tendermint/go-amino => github.com/binance-chain/bnc-go-amino v0.14.1-binance.1
