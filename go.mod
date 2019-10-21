module gitlab.com/thorchain/bepswap/statechain

go 1.13

require (
	github.com/binance-chain/go-sdk v1.1.3
	github.com/btcsuite/btcd v0.0.0-20190824003749-130ea5bddde3 // indirect
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/cosmos/cosmos-sdk v0.37.3
	github.com/didip/tollbooth v4.0.2+incompatible
	github.com/go-resty/resty/v2 v2.0.0
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/gorilla/mux v1.7.0
	github.com/mattn/go-isatty v0.0.7 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.2.1 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.32.6
	github.com/tendermint/tm-db v0.2.0
	gitlab.com/thorchain/bepswap/common v1.0.1-0.20191019041734-5bea39d3cf49
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550 // indirect
	golang.org/x/net v0.0.0-20191014212845-da9a3fd4c582 // indirect
	golang.org/x/sys v0.0.0-20191020212454-3e7259c5e7c2 // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/genproto v0.0.0-20191009194640-548a555dbc03 // indirect
	google.golang.org/grpc v1.24.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
	gopkg.in/yaml.v2 v2.2.4 // indirect
)

replace gitlab.com/thorchain/bepswap/statechain => ../statechain
