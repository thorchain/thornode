include Makefile.cicd
.PHONY: test export

GOBIN?=${GOPATH}/bin

all: lint install

install: go.sum
	GO111MODULE=on go install -tags "$(build_tags)" ./cmd/sscli
	GO111MODULE=on go install -tags "$(build_tags)" ./cmd/ssd
	GO111MODULE=on go install -v ./cmd/observed
	GO111MODULE=on go install -v ./cmd/signd

tools: install
	GO111MODULE=on go install -tags "$(build_tags)" ./tools/smoke
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
	ssd start

start-rest:
	sscli rest-server

setup: install
	./scripts/setup.sh

reset: clean 
	make install
	./scripts/reset.sh

clean:
	rm -rf ~/.ssd
	rm -rf ~/.sscli
	rm -f ${GOBIN}/{smoke,generate,sweep,sscli,ssd,observe,signd}

.envrc: install
	@generate -t MASTER > .envrc
	@generate -t POOL >> .envrc

extract: tools
	@extract -f "${FILE}" -p "${PASSWORD}" -t ${TYPE}

smoke-test-audit: tools
	@smoke -f ${FAUCET_KEY} -p ${BINANCE_PRIVATE_KEY} -c test/smoke/definitions/full/smoke-test-audit.json -e ${ENV}

smoke-test-refund: tools
	@smoke -f ${FAUCET_KEY} -p ${BINANCE_PRIVATE_KEY} -c test/smoke/definitions/full/smoke-test-refund.json -e ${ENV}

gas: tools
	@smoke -f ${FAUCET_KEY} -p ${BINANCE_PRIVATE_KEY} -c test/smoke/definitions/unit/gas.json -e ${ENV}

stake: tools
	@smoke -f ${FAUCET_KEY} -p ${BINANCE_PRIVATE_KEY} -c test/smoke/definitions/unit/stake.json -e ${ENV}

swap: tools
	@smoke -f ${FAUCET_KEY} -p ${BINANCE_PRIVATE_KEY} -c test/smoke/definitions/unit/swap.json -e ${ENV}

sweep: tools
	@sweep -m ${FAUCET_KEY} -k ${BINANCE_PRIVATE_KEY} -d true

export:
	ssd export
