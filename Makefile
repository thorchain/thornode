include Makefile.cicd
.PHONY: test tools export healthcheck

GOBIN?=${GOPATH}/bin

# variables default for healthcheck against full stack in docker
CHAIN_API?=localhost:1317
MIDGARD_API?=localhost:8080

all: lint install

install: go.sum
	go install -tags "${TAG}" ./cmd/thorcli
	go install -tags "${TAG}" ./cmd/thord
	go install ./cmd/bifrost

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
	@go test -v -tags testnet -coverprofile cover.txt ./...

coverage-report: test-coverage
	@go tool cover -html=cover.txt

clear:
	clear

test:
	@go test -tags testnet ./...

test-watch: clear
	@gow -c test -tags testnet -mod=readonly ./...

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

build:
	@go build -tags "${TAG}" ./...

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
