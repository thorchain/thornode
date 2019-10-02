include Makefile.ledger
all: lint install

install: go.sum
		GO111MODULE=on go install -tags "$(build_tags)" ./cmd/sscli
		GO111MODULE=on go install -tags "$(build_tags)" ./cmd/ssd
		GO111MODULE=on go install -tags "$(build_tags)" ./cmd/smoke

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify

lint:
	@golangci-lint run --deadline=15m
	@go mod verify

test-coverage:
	@go test -v -coverprofile .testCoverage.txt ./...

test:
	@go test ./...

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
	rm ${GOBIN}/smoke
	ssd unsafe-reset-all

export:
	ssd export
