
setup:
	@bundle install

test-internal:
	@rspec --fail-fast -f documentation

wait-for-blockchain:
	@./wait-for-blockchain.sh

install:
	make -C ../.. setup

run:
	NET=testnet make -C ../.. start-daemon > /dev/stdout 2>&1 &
	NET=testnet make -C ../.. start-rest > /var/log/rest.log 2>&1 &

start-daemon:
	make -C ../.. reset start-daemon

start-rest:
	make -C ../.. start-rest

export: stopchain
	thord export > genesis.json

validate-genesis:
	thord validate-genesis genesis.json

stopchain:
	killall thord

sh:
	docker run -it --rm -v ${GOPATH}:/root/go -v $(shell pwd)/../..:/root/go/src/gitlab.com/thorchain/bepswap/thornode -w /root/go/src/gitlab.com/thorchain/bepswap/thornode -e GOPATH=/root/go -e GOBIN=/go/bin -e NET=oops cbarraford/golang-ruby /bin/sh

test:
	docker run --rm -v ${GOPATH}:/root/go -v $(shell pwd)/../..:/root/go/src/gitlab.com/thorchain/bepswap/thornode -w /root/go/src/gitlab.com/thorchain/bepswap/thornode -e GOPATH=/root/go -e GOBIN=/go/bin -e NET=oops  cbarraford/golang-ruby make -C ./test/integration install setup run wait-for-blockchain test-internal export validate-genesis