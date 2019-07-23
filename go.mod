module github.com/jpthor/cosmos-swap

go 1.12

require (
	github.com/binance-chain/go-sdk v1.0.7
	github.com/cosmos/cosmos-sdk v0.35.0
	github.com/gorilla/mux v1.7.0
	github.com/mattn/go-isatty v0.0.7 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/procfs v0.0.0-20190328153300-af7bedc223fb // indirect
	github.com/rs/zerolog v1.14.3
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.0.3
	github.com/syndtr/goleveldb v1.0.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/iavl v0.12.2 // indirect
	github.com/tendermint/tendermint v0.31.5
	golang.org/x/sys v0.0.0-20190329044733-9eb1bfa1ce65 // indirect
	google.golang.org/appengine v1.4.0 // indirect
	google.golang.org/genproto v0.0.0-20190327125643-d831d65fe17d // indirect
)

replace golang.org/x/crypto => github.com/johnnyluo/crypto v0.0.0-20190722223544-3f5ecfe86f08

replace github.com/jpthor/cosmos-swap => ../cosmos-swap
