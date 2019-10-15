
setup:
	@bundle install

test-internal:
	@rspec --fail-fast -f documentation

wait-for-blockchain:
	@./wait-for-blockchain.sh

install:
	make -C ../.. setup

run:
	make -C ../.. start-daemon > /var/log/daemon.log 2>&1 &
	make -C ../.. start-rest > /var/log/rest.log 2>&1 &

start-daemon:
	make -C ../.. reset start-daemon

start-rest:
	make -C ../.. start-rest

export: stopchain
	ssd export > genesis.json

validate-genesis:
	ssd validate-genesis genesis.json

stopchain:
	killall ssd

sh:
	docker run -it --rm -v ${GOPATH}:/root/go -v $(shell pwd)/../..:/root/go/src/gitlab.com/thorchain/bepswap/statechain -w /root/go/src/gitlab.com/thorchain/bepswap/statechain -e GOPATH=/root/go -e GOBIN=/go/bin cbarraford/golang-ruby /bin/sh

test:
	docker run --rm -v ${GOPATH}:/root/go -v $(shell pwd)/../..:/root/go/src/gitlab.com/thorchain/bepswap/statechain -w /root/go/src/gitlab.com/thorchain/bepswap/statechain -e GOPATH=/root/go -e GOBIN=/go/bin cbarraford/golang-ruby make -C ./tests/integration install setup run wait-for-blockchain test-internal export validate-genesis
