#
# Build
#
FROM golang:1.13 AS build

WORKDIR /go/src/app

COPY . .

ENV GOBIN=/go/bin
ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go mod verify
RUN go get -d -v ./...

RUN go build -a -installsuffix cgo -o observed ./cmd/observed
RUN go build -a -installsuffix cgo -o thorcli ./cmd/thorcli

#
# Observer
#
FROM alpine

RUN apk add --update jq curl && \
    rm -rf /var/cache/apk/*

ENV PATH="${PATH}:/go/bin"

# Copy the compiled binaires over.
COPY --from=build /go/src/app/observed /go/bin/
COPY --from=build /go/src/app/thorcli /go/bin/

CMD ["observed", "-c", "/etc/observe/observed/config.json"]
