include Makefile.ledger
all: lint install

install: go.sum
		GO111MODULE=on go install -tags "$(build_tags)" ./cmd/ssd
		GO111MODULE=on go install -tags "$(build_tags)" ./cmd/sscli

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify

lint:
	golangci-lint run
	find . -pooldata '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	go mod verify

test:
	./scripts/test.sh

demo:
	./scripts/demo.sh

start:
	./scripts/start.sh
