module gitlab.com/thorchain/bepswap/observe

go 1.12

require (
	github.com/binance-chain/go-sdk v1.0.8
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cosmos/cosmos-sdk v0.37.0
	github.com/golang/mock v1.3.1 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/gorilla/websocket v1.4.1
	github.com/hashicorp/go-retryablehttp v0.5.4
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/onsi/ginkgo v1.9.0 // indirect
	github.com/onsi/gomega v1.6.0 // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0
	github.com/rakyll/statik v0.1.6 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563 // indirect
	github.com/rs/cors v1.7.0 // indirect
	github.com/rs/zerolog v1.14.3
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20190318030020-c3a204f8e965
	github.com/tendermint/crypto v0.0.0-20190823183015-45b1026d81ae // indirect
	gitlab.com/thorchain/bepswap/common v1.0.0
	gitlab.com/thorchain/bepswap/statechain v0.0.0-20190914112626-edde6ee51ae0
	golang.org/x/sys v0.0.0-20190913121621-c3b328c6e5a7 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
	gopkg.in/resty.v1 v1.12.0
)

replace gitlab.com/thorchain/bepswap/observe => ../observe

replace github.com/tendermint/go-amino => github.com/binance-chain/bnc-go-amino v0.14.1-binance.1
