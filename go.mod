module gitlab.com/thorchain/bepswap/observe

go 1.12

require (
	github.com/binance-chain/go-sdk v1.1.3
	github.com/btcsuite/btcd v0.0.0-20190926002857-ba530c4abb35 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cosmos/cosmos-sdk v0.37.2
	github.com/golang/mock v1.3.1 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.5.4
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/onsi/ginkgo v1.9.0 // indirect
	github.com/onsi/gomega v1.6.0 // indirect
	github.com/pelletier/go-toml v1.5.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/procfs v0.0.5 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563 // indirect
	github.com/rs/zerolog v1.14.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	github.com/syndtr/goleveldb v1.0.1-0.20190318030020-c3a204f8e965
	github.com/tendermint/crypto v0.0.0-20190823183015-45b1026d81ae // indirect
	gitlab.com/thorchain/bepswap/common v1.0.0
	gitlab.com/thorchain/bepswap/statechain v0.0.0-20191006024012-f2356437f55f
	golang.org/x/crypto v0.0.0-20191002192127-34f69633bfdc // indirect
	golang.org/x/net v0.0.0-20191007182048-72f939374954 // indirect
	golang.org/x/sys v0.0.0-20191008105621-543471e840be // indirect
	google.golang.org/genproto v0.0.0-20191007204434-a023cd5227bd // indirect
	google.golang.org/grpc v1.24.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
	gopkg.in/resty.v1 v1.12.0
)

replace gitlab.com/thorchain/bepswap/observe => ../observe

replace github.com/tendermint/go-amino => github.com/binance-chain/bnc-go-amino v0.14.1-binance.1
