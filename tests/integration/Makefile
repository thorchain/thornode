
setup:
	@bundle install

test-internal:
	@rspec --fail-fast

wait-for-blockchain:
	@./wait-for-blockchain.sh

install:
	make -C ../.. setup

run:
	make -C ../.. start-daemon > /var/log/daemon.log 2>&1 &
	make -C ../.. start-rest > /var/log/rest.log 2>&1 &

test:
	docker run --rm -v ${GOPATH}:/root/go -v $(shell pwd)/../..:/root/go/src/gitlab.com/thorchain/statechain -w /root/go/src/gitlab.com/thorchain/statechain -e GOPATH=/root/go -e GOBIN=/go/bin cbarraford/golang-ruby make -C ./tests/integration install setup run wait-for-blockchain test-internal
