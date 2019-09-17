include Makefile.ledger
all: lint install

install: go.sum
		GO111MODULE=on go install -tags "$(build_tags)" ./cmd/sscli
		GO111MODULE=on go install -tags "$(build_tags)" ./cmd/ssd

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
	@go build -o build/ssd ./cmd/ssd
	@go build -o build/sscli ./cmd/sscli

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
	ssd unsafe-reset-all

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build
#########################################################
### Local validator nodes using docker and docker-compose
build-docker-statechainnode:
	$(MAKE) -C networks/local

# Run a 4-node testnet locally
localnet-start: build-linux localnet-stop
	@if ! [ -f build/node0/ssd/config/genesis.json ]; then docker run --rm -v $(CURDIR)/cmd:/ssd:Z thorchain/statechainnode testnet --v 4 -o . --starting-ip-address 192.168.10.2 ; fi
	docker-compose up -d

# Stop testnet
localnet-stop:
	docker-compose down