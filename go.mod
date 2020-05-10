module gitlab.com/thorchain/thornode

go 1.13

require (
	github.com/binance-chain/go-sdk v1.2.2
	github.com/binance-chain/ledger-cosmos-go v0.9.9 // indirect
	github.com/binance-chain/tss-lib v1.3.1
	github.com/blang/semver v3.5.1+incompatible
	github.com/btcsuite/btcd v0.20.1-beta.0.20200414114020-8b54b0b96418
	github.com/btcsuite/btcutil v1.0.2
	github.com/cosmos/cosmos-sdk v0.37.7
	github.com/cosmos/ledger-cosmos-go v0.11.1 // indirect
	github.com/didip/tollbooth v4.0.2+incompatible
	github.com/ethereum/go-ethereum v1.9.12
	github.com/go-kit/kit v0.10.0 // indirect
	github.com/golang/protobuf v1.3.4 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/go-retryablehttp v0.6.4
	github.com/ipfs/go-datastore v0.4.4 // indirect
	github.com/ipfs/go-log v1.0.2
	github.com/libp2p/go-libp2p-kad-dht v0.5.0 // indirect
	github.com/multiformats/go-multiaddr v0.2.1
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/prometheus/client_golang v1.5.0
	github.com/prometheus/procfs v0.0.10 // indirect
	github.com/rakyll/statik v0.1.6 // indirect
	github.com/rs/zerolog v1.18.0
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20190923125748-758128399b1d
	github.com/tendermint/btcd v0.1.1
	github.com/tendermint/crypto v0.0.0-20191022145703-50d29ede1e15 // indirect
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.32.9
	github.com/tendermint/tm-db v0.2.0
	github.com/zondax/ledger-go v0.11.0 // indirect
	gitlab.com/thorchain/tss/go-tss v0.0.0-20200510003725-b211cb28c534
	gitlab.com/thorchain/txscript v0.0.0-20200413023754-8aaf3443d92b
	go.uber.org/multierr v1.5.0 // indirect
	golang.org/x/crypto v0.0.0-20200414173820-0848c9571904 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/tools v0.0.0-20200409210453-700752c24408 // indirect
	google.golang.org/genproto v0.0.0-20200304201815-d429ff31ee6c // indirect
	google.golang.org/grpc v1.27.1 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15
	gopkg.in/ini.v1 v1.52.0 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace github.com/tendermint/go-amino => github.com/binance-chain/bnc-go-amino v0.14.1-binance.1
