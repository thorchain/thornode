module gitlab.com/thorchain/bepswap/observe

go 1.12

require (
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/binance-chain/go-sdk v1.0.8
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/gorilla/websocket v1.4.0
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/kr/logfmt v0.0.0-20140226030751-b84e30acd515 // indirect
	github.com/rs/zerolog v1.14.3
)

replace gitlab.com/thorchain/bepswap/observe => ../observe

replace github.com/tendermint/go-amino => github.com/binance-chain/bnc-go-amino v0.14.1-binance.1
