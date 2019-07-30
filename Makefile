all: lint install

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify

lint:
	golangci-lint run
	find . -pooldata '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	go mod verify

build:
	go build

install:
	go install -v

test:
	./scripts/test.sh

start:
	./scripts/start.sh

clean:
	rm ${GOBIN}/observe
