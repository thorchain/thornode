include Makefile.cicd
.PHONY: test export

GOBIN?=${GOPATH}/bin

all: lint install

install: go.sum
	GO111MODULE=on go install -tags "$(build_tags)" ./cmd/thorcli
	GO111MODULE=on go install -tags "$(build_tags)" ./cmd/thord
	GO111MODULE=on go install -v ./cmd/observed
	GO111MODULE=on go install -v ./cmd/signd

tools: install
	GO111MODULE=on go install -tags "$(build_tags)" ./tools/bsinner
	GO111MODULE=on go install -tags "$(build_tags)" ./tools/generate
	GO111MODULE=on go install -tags "$(build_tags)" ./tools/extract
	GO111MODULE=on go install -tags "$(build_tags)" ./tools/sweep

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	GO111MODULE=on go mod verify

test-coverage:
	@go test -v -coverprofile .testCoverage.txt ./...

coverage-report: test-coverage
	@tool cover -html=.testCoverage.txt

clear:
	clear

test:
	@go test -mod=readonly ./...

test-watch: clear
	@./scripts/watch.bash

lint-pre:
	@test -z $(gofmt -l .) # checks code is in proper format
	@go mod verify

lint: lint-pre
	@golangci-lint run
	@go mod verify

lint-verbose: lint-pre
	@golangci-lint run -v

build:
	@go build ./...

start-observe:
	observe

start-daemon:
	thord start

start-rest:
	thorcli rest-server

setup: install
	./build/scripts/localdev.sh

reset: clean install
	./build/scripts/localdev.sh

clean:
	rm -rf ~/.thord
	rm -rf ~/.thorcli
	rm -f ${GOBIN}/{bsinner,generate,sweep,thorcli,thord,observe,signd}

.envrc: install
	@generate -t MASTER > .envrc
	@generate -t POOL >> .envrc

extract: tools
	@extract -f "${FILE}" -p "${PASSWORD}" -t ${TYPE}

smoke-test-audit: tools
	@bsinner -f ${FAUCET_KEY} -p ${BINANCE_PRIVATE_KEY} -c test/smoke/definitions/full/smoke-test-audit.json -e ${ENV}

smoke-test-refund: tools
	@bsinner -f ${FAUCET_KEY} -p ${BINANCE_PRIVATE_KEY} -c test/smoke/definitions/full/smoke-test-refund.json -e ${ENV}

gas: tools
	@bsinner -f ${FAUCET_KEY} -p ${BINANCE_PRIVATE_KEY} -c test/smoke/definitions/unit/gas.json -e ${ENV}

stake: tools
	@bsinner -f ${FAUCET_KEY} -p ${BINANCE_PRIVATE_KEY} -c test/smoke/definitions/unit/stake.json -e ${ENV}

swap: tools
	@bsinner -f ${FAUCET_KEY} -p ${BINANCE_PRIVATE_KEY} -c test/smoke/definitions/unit/swap.json -e ${ENV}

sweep: tools
	@sweep -m ${FAUCET_KEY} -k ${BINANCE_PRIVATE_KEY} -d true

export:
	thord export
