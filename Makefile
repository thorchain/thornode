include Makefile.cicd
.PHONY: build test tools export healthcheck

GOBIN?=${GOPATH}/bin
NOW=$(shell date +'%Y-%m-%d_%T')
DEFAULT_BUILD_ARGS=-ldflags "-X 'gitlab.com/thorchain/thornode/constants.Version=0.1.0' -X 'gitlab.com/thorchain/thornode/constants.GitCommit=$(shell git rev-parse --short HEAD)' -X gitlab.com/thorchain/thornode/constants.BuildTime=${NOW}" -tags "${TAG} -a -installsuffix cgo"

BINARIES=./cmd/thorcli ./cmd/thord ./cmd/bifrost ./tools/generate

# variables default for healthcheck against full stack in docker
CHAIN_API?=localhost:1317
MIDGARD_API?=localhost:8080

all: lint install

build:
	go build ${DEFAULT_BUILD_ARGS} ${BUILD_ARGS} ${BINARIES}

install: go.sum
	go install ${DEFAULT_BUILD_ARGS} ${BUILD_ARGS} ${BINARIES}

install-testnet:
	TAG=testnet make install

install-mocknet:
	TAG=mocknet make install

tools:
	go install ./tools/generate
	go install ./tools/extract

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	go mod verify

test-coverage:
	@TAG=testnet
	@go test ${DEFAULT_BUILD_ARGS} -v -coverprofile cover.txt ./...

coverage-report: test-coverage
	@go tool cover -html=cover.txt

clear:
	clear

test:
	@TAG=testnet
	@go test ${DEFAULT_BUILD_ARGS} ./...

test-watch: clear
	@TAG=tesnet
	@gow -c test ${DEFAULT_BUILD_ARGS} -mod=readonly ./...

format:
	@gofumpt -w .

lint-pre:
	@gofumpt -l cmd x bifrost common constants tools # for display
	@test -z "$(shell gofumpt -l cmd x bifrost common constants tools)" # cause error
	@go mod verify

lint: lint-pre
	@golangci-lint run

lint-verbose: lint-pre
	@golangci-lint run -v

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
	rm -f ${GOBIN}/{generate,thorcli,thord,bifrost}

.envrc: install
	@generate -t MASTER > .envrc
	@generate -t POOL >> .envrc

extract: tools
	@extract -f "${FILE}" -p "${PASSWORD}" -t ${TYPE}

# updates our tss dependency
tss:
	go get gitlab.com/thorchain/tss/go-tss

export:
	thord export

pull:
	docker pull registry.gitlab.com/thorchain/thornode:mocknet
	docker pull registry.gitlab.com/thorchain/midgard
	docker pull registry.gitlab.com/thorchain/bepswap/bepswap-web-ui
	docker pull registry.gitlab.com/thorchain/bepswap/mock-binance
