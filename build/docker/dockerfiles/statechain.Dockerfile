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

RUN go build -a -installsuffix cgo -o generate ./tools/generate
RUN go build -a -installsuffix cgo -o thord ./cmd/thord
RUN go build -a -installsuffix cgo -o thorcli ./cmd/thorcli

#
# Statechain
#
FROM alpine

RUN apk add --update jq curl nginx && \
    rm -rf /var/cache/apk/*

ENV PATH="${PATH}:/go/bin"

# Copy the compiled binaires over.
COPY --from=build /go/src/app/generate /go/bin/
COPY --from=build /go/src/app/thord /go/bin/
COPY --from=build /go/src/app/thorcli /go/bin/

# Add users.
RUN adduser -Ds /bin/sh www-data -G www-data

# TODO Move away from needing nginx
# Setup Nginx.
ADD etc/nginx/nginx.conf /etc/nginx/

EXPOSE 9000
EXPOSE 1317
EXPOSE 81
