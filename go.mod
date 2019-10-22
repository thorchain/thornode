module gitlab.com/thorchain/bepswap/statechain

go 1.13

require (
	github.com/binance-chain/go-sdk v1.1.3
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cosmos/cosmos-sdk v0.37.3
	github.com/didip/tollbooth v4.0.2+incompatible
	github.com/go-resty/resty/v2 v2.0.0
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/go-retryablehttp v0.6.2
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.2.1
	github.com/rs/zerolog v1.15.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	github.com/syndtr/goleveldb v1.0.1-0.20190318030020-c3a204f8e965
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.32.6
	github.com/tendermint/tm-db v0.2.0
	gitlab.com/thorchain/bepswap/common v1.0.1
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
	gopkg.in/resty.v1 v1.12.0
)

replace gitlab.com/thorchain/bepswap/statechain => ../statechain

replace github.com/tendermint/go-amino => github.com/binance-chain/bnc-go-amino v0.14.1-binance.1
