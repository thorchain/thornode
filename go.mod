module gitlab.com/thorchain/bepswap/statechain

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.36.0-rc1
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d // indirect
	github.com/gorilla/mux v1.7.0
	github.com/mattn/go-isatty v0.0.7 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/procfs v0.0.0-20190328153300-af7bedc223fb // indirect
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.3.2
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.32.1
	gitlab.com/thorchain/bepswap/common v0.0.0-20190816093251-b84c21cee45c
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/sys v0.0.0-20190329044733-9eb1bfa1ce65 // indirect
	google.golang.org/appengine v1.4.0 // indirect
	google.golang.org/genproto v0.0.0-20190327125643-d831d65fe17d // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
)

replace golang.org/x/crypto => github.com/johnnyluo/crypto v0.0.0-20190722223544-3f5ecfe86f08

replace gitlab.com/thorchain/bepswap/statechain => ../statechain
