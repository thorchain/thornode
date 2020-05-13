include Makefile.cicd
.PHONY: build test tools export healthcheck

GOBIN?=${GOPATH}/bin
NOW=$(shell date +'%Y-%m-%d_%T')
COMMIT:=$(shell git log -1 --format='%H')
VERSION:=$(shell cat version)
TAG?=testnet

ldflags = -X gitlab.com/thorchain/thornode/constants.Version=$(VERSION) \
		  -X gitlab.com/thorchain/thornode/constants.GitCommit=$(COMMIT) \
		  -X gitlab.com/thorchain/thornode/constants.BuildTime=${NOW} \
		  -X github.com/cosmos/cosmos-sdk/version.Name=THORChain \
	      -X github.com/cosmos/cosmos-sdk/version.ServerName=thord \
		  -X github.com/cosmos/cosmos-sdk/version.ClientName=thorcli \
	      -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	      -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X github.com/cosmos/cosmos-sdk/version.BuildTags=$(TAG)

BUILD_FLAGS := -ldflags '$(ldflags)' -tags ${TAG} -a

BINARIES=./cmd/thorcli ./cmd/thord ./cmd/bifrost ./tools/generate

# variables default for healthcheck against full stack in docker
CHAIN_API?=localhost:1317
MIDGARD_API?=localhost:8080

all: lint install

build:
	go build ${BUILD_FLAGS} ${BINARIES}

install: go.sum
	go install ${BUILD_FLAGS} ${BINARIES}

tools:
	go install ./tools/generate
	go install ./tools/extract

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	go mod verify

test-coverage:
	@go test ${BUILD_FLAGS} -v -coverprofile coverage.out ./...

coverage-report: test-coverage
	@go tool cover -html=cover.txt

clear:
	clear

test:
	@go test ${BUILD_FLAGS} ./...

test-watch: clear
	@gow -c test ${BUILD_FLAGS} -mod=readonly ./...

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
