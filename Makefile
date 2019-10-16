GOBIN?=${GOPATH}/bin

all: lint install

install: go.sum
	GO111MODULE=on go install -tags "$(build_tags)" ./cmd/sscli
	GO111MODULE=on go install -tags "$(build_tags)" ./cmd/ssd

tools: install
	GO111MODULE=on go install -tags "$(build_tags)" ./tools/smoke
	GO111MODULE=on go install -tags "$(build_tags)" ./tools/generate
	GO111MODULE=on go install -tags "$(build_tags)" ./tools/extract
	GO111MODULE=on go install -tags "$(build_tags)" ./tools/sweep

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

setup: install
	./scripts/setup.sh

reset: clean 
	make install
	./scripts/reset.sh

clean:
	rm -rf ~/.ssd
	rm -rf ~/.sscli
	ssd unsafe-reset-all
	rm ${GOBIN}/{smoke,generate,sweep,sscli,ssd}

export:
	ssd export

.envrc: install
	@generate -t MASTER > .envrc
	@generate -t POOL >> .envrc

extract: tools
	@extract -f "${FILE}" -p "${PASSWORD}" -t ${TYPE}

smoke-test-audit: tools
	@smoke -f ${FAUCET_KEY} -p ${POOL_KEY} -c test/smoke/definitions/full/smoke-test-audit.json -e ${ENV}

smoke-test-refund: tools
	@smoke -f ${FAUCET_KEY} -p ${POOL_KEY} -c test/smoke/definitions/full/smoke-test-refund.json -e ${ENV}

seed: tools
	@smoke -f ${FAUCET_KEY} -p ${POOL_KEY} -c test/smoke/definitions/unit/seed.json -e ${ENV}

gas: tools
	@smoke -f ${FAUCET_KEY} -p ${POOL_KEY} -c test/smoke/definitions/unit/gas.json -e ${ENV}

stake: tools
	@smoke -f ${FAUCET_KEY} -p ${POOL_KEY} -c test/smoke/definitions/unit/stake.json -e ${ENV}

swap: tools
	@smoke -f ${FAUCET_KEY} -p ${POOL_KEY} -c test/smoke/definitions/unit/swap.json -e ${ENV}

sweep: tools
	@sweep -m ${MASTER_KEY} -k ${KEY_LIST}
