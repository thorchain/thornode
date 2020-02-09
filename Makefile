include Makefile.cicd
.PHONY: test tools export healthcheck

GOBIN?=${GOPATH}/bin

# variables default for healthcheck against full stack in docker
CHAIN_API?=localhost:1317
MIDGARD_API?=localhost:8080

all: lint install

install: go.sum
	go install -tags "${TAGS}" ./cmd/thorcli
	go install -tags "${TAGS}" ./cmd/thord
	go install ./cmd/bifrost

install-testnet:
	TAGS=testnet make install

install-sandbox:
	TAGS=sandbox make install

tools:
	go install ./tools/bsinner
	go install ./tools/generate
	go install ./tools/extract
	go install ./tools/sweep

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	go mod verify

test-coverage:
	@go test -v -tags testnet -coverprofile cover.txt ./...

coverage-report: test-coverage
	@go tool cover -html=cover.txt

clear:
	clear

test:
	@go test -tags testnet ./...

test-watch: clear
	@gow -c test -tags testnet -mod=readonly ./...

lint-pre:
	@test -z "$(shell gofumpt -l .)"
	@go mod verify

lint: lint-pre
	@golangci-lint run

lint-verbose: lint-pre
	@golangci-lint run -v

build:
	@go build -tags "${TAGS}" ./...

start-daemon:
	thord start --log_level "main:info,state:debug,*:error"

start-rest:
	thorcli rest-server

setup: install
	./build/scripts/localdev.sh

reset: clean install
	./build/scripts/localdev.sh

clean:
	rm -rf ~/.thor*
	rm -f ${GOBIN}/{bsinner,generate,sweep,thorcli,thord,bifrost}

.envrc: install
	@generate -t MASTER > .envrc
	@generate -t POOL >> .envrc

extract: tools
	@extract -f "${FILE}" -p "${PASSWORD}" -t ${TYPE}

sweep: tools
	@sweep -m ${FAUCET_KEY} -k ${PRIV_KEY} -d true

smoke-test: tools install
	./build/scripts/smoke.sh

smoke-local: smoke-standalone

smoke-standalone:
	make -C build/docker reset-mocknet-standalone
	bsinner -a localhost:26660 -b ./test/smoke/scenarios/standalone/balances.json -t ./test/smoke/scenarios/standalone/transactions.json -e local -x -g

smoke-genesis:
	make -C build/docker reset-mocknet-genesis
	bsinner -a localhost:26660 -b ./test/smoke/scenarios/genesis/balances.json -t ./test/smoke/scenarios/genesis/transactions.json -e local -x -g

healthcheck:
	@CHAIN_API=${CHAIN_API} MIDGARD_API=${MIDGARD_API} go test -tags healthcheck ./tools/healthcheck/...

export:
	thord export

pull:
	docker pull registry.gitlab.com/thorchain/thornode
	docker pull registry.gitlab.com/thorchain/tss/go-tss
	docker pull registry.gitlab.com/thorchain/midgard
	docker pull registry.gitlab.com/thorchain/bepswap/bepswap-react-app
	docker pull registry.gitlab.com/thorchain/bepswap/mock-binance
