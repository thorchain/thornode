module gitlab.com/thorchain/bepswap/observe

go 1.12

require (
	cloud.google.com/go v0.44.3 // indirect
	github.com/binance-chain/go-sdk v1.0.8
	github.com/btcsuite/btcd v0.0.0-20190824003749-130ea5bddde3 // indirect
	github.com/btcsuite/goleveldb v1.0.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.15+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/cosmos/cosmos-sdk v0.37.0
	github.com/go-kit/kit v0.9.0 // indirect
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/google/pprof v0.0.0-20190723021845-34ac40c74b70 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/gorilla/websocket v1.4.1
	github.com/grpc-ecosystem/grpc-gateway v1.9.6 // indirect
	github.com/hashicorp/go-retryablehttp v0.5.4
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/kisielk/errcheck v1.2.0 // indirect
	github.com/kkdai/bstream v1.0.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/pty v1.1.8 // indirect
	github.com/libp2p/go-buffer-pool v0.0.2 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/onsi/ginkgo v1.9.0 // indirect
	github.com/onsi/gomega v1.6.0 // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/prometheus/procfs v0.0.4 // indirect
	github.com/rakyll/statik v0.1.6 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563 // indirect
	github.com/rogpeppe/fastuuid v1.2.0 // indirect
	github.com/rs/cors v1.7.0 // indirect
	github.com/rs/zerolog v1.14.3
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20190318030020-c3a204f8e965
	github.com/tendermint/crypto v0.0.0-20190823183015-45b1026d81ae // indirect
	github.com/ugorji/go v1.1.7 // indirect
	gitlab.com/thorchain/bepswap/common v0.0.0-20190823123750-2e16dc69db55
	gitlab.com/thorchain/bepswap/statechain v0.0.0-20190826134211-8df2518d6572
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586 // indirect
	golang.org/x/image v0.0.0-20190823064033-3a9bac650e44 // indirect
	golang.org/x/mobile v0.0.0-20190826170111-cafc553e1ac5 // indirect
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7 // indirect
	golang.org/x/sys v0.0.0-20190826190057-c7b8b68b1456 // indirect
	golang.org/x/tools v0.0.0-20190826234050-71894ab67ee3 // indirect
	google.golang.org/api v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55 // indirect
	google.golang.org/grpc v1.23.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
	gopkg.in/resty.v1 v1.12.0
	honnef.co/go/tools v0.0.1-2019.2.2 // indirect
)

replace gitlab.com/thorchain/bepswap/observe => ../observe

replace github.com/tendermint/go-amino => github.com/binance-chain/bnc-go-amino v0.14.1-binance.1
