include Makefile.cicd
.PHONY: test tools export

GOBIN?=${GOPATH}/bin

all: lint install

install: go.sum
	go install ./cmd/thorcli
	go install ./cmd/thord
	go install ./cmd/bifrost

tools:
	go install ./tools/bsinner
	go install ./tools/generate
	go install ./tools/extract
	go install ./tools/sweep

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	go mod verify

test-coverage:
	@go test -v -coverprofile .testCoverage.txt ./...

coverage-report: test-coverage
	@tool cover -html=.testCoverage.txt

clear:
	clear

test:
	@go test ./...

test-watch: clear
	@gow -c test -mod=readonly ./...

lint-pre:
	@test -z $(gofmt -l .) # checks code is in proper format
	@go mod verify

lint: lint-pre
	@golangci-lint run

lint-verbose: lint-pre
	@golangci-lint run -v

build:
	@go build ./...

start-daemon:
	thord start --log_level "main:info,state:debug,*:error"

start-rest:
	thorcli rest-server

setup: install
	./build/scripts/localdev.sh

reset: clean install
	./build/scripts/localdev.sh

clean:
	rm -rf ~/.thord
	rm -rf ~/.thorcli
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

export:
	thord export
