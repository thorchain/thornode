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

start-daemon-compose:
	ssd start --rpc.laddr "tcp://0.0.0.0:26657"

start-rest-compose:
	sscli rest-server --laddr "tcp://0.0.0.0:1317"  --node "tcp://daemon:26657" --chain-id sschain --trust-node

setup:
	./scripts/setup.sh

reset: clean
	./scripts/reset.sh

clean:
	rm -rf ~/.ssd
	ssd unsafe-reset-all
