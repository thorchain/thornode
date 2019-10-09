include Makefile.ledger

all: lint install

install: go.sum
	GO111MODULE=on go install -tags "$(build_tags)" ./cmd/sscli
	GO111MODULE=on go install -tags "$(build_tags)" ./cmd/ssd
	GO111MODULE=on go install -tags "$(build_tags)" ./cmd/smoke
	GO111MODULE=on go install -tags "$(build_tags)" ./cmd/generate
	GO111MODULE=on go install -tags "$(build_tags)" ./cmd/extract
	GO111MODULE=on go install -tags "$(build_tags)" ./cmd/sweep

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	GO111MODULE=on go mod verify

lint:
	@golangci-lint run --deadline=15m
	@go mod verify

test-coverage:
	@go test -mod=readonly -v -coverprofile .testCoverage.txt ./...

test:
	@go test -mod=readonly ./...

clear:
	clear

test-watch: clear
	@./scripts/watch.bash

build:
	@go build ./...

start: install start-daemon

start-daemon:
	ssd start

start-rest:
	sscli rest-server

setup:
	./scripts/setup.sh

reset: clean
	./scripts/reset.sh

clean:
	rm -rf ~/.ssd
	rm ${GOBIN}/{smoke,generate,sweep}
	ssd unsafe-reset-all

export:
	ssd export

.envrc: install
	@generate -t MASTER > .envrc
	@generate -t POOL >> .envrc

extract: install
	@extract -f "${FILE}" -p "${PASSWORD}" -t ${TYPE}

smoke-test-audit: install
	@smoke -b ${BANK_KEY} -p ${POOL_KEY} -c tests/smoke/smoke-test-audit.json -e ${ENV} -d ${DEBUG}

smoke-test-refund: install
	@smoke -b ${BANK_KEY} -p ${POOL_KEY} -c tests/smoke/smoke-test-refund.json -e ${ENV} -d ${DEBUG}

seed: install
	@smoke -b ${BANK_KEY} -p ${POOL_KEY} -c tests/unit/seed.json -e ${ENV} -d ${DEBUG}

gas: install
	@smoke -b ${BANK_KEY} -p ${POOL_KEY} -c tests/unit/gas.json -e ${ENV} -d ${DEBUG}

stake: install
	@smoke -b ${BANK_KEY} -p ${POOL_KEY} -c tests/unit/stake.json -e ${ENV} -d ${DEBUG}

swap: install
	@smoke -b ${BANK_KEY} -p ${POOL_KEY} -c tests/unit/swap.json -e ${ENV} -d ${DEBUG}

sweep: install
	@sweep -m ${MASTER_KEY} -k ${KEY_LIST} -d ${DEBUG}
