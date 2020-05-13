#
# THORChain
#

#
# Build
#
FROM golang:1.13 AS build

ENV GOBIN=/go/bin
ENV GOPATH=/go
ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TAG=mainnet
RUN make install

#
# Main
#
FROM alpine

RUN apk add --update jq curl nginx && \
    rm -rf /var/cache/apk/*

# Copy the compiled binaires over.
COPY --from=build /go/bin/generate /usr/bin/
COPY --from=build /go/bin/thord /usr/bin/
COPY --from=build /go/bin/thorcli /usr/bin/
COPY --from=build /go/bin/bifrost /usr/bin/

# Add users.
RUN adduser -Ds /bin/sh www-data -G www-data

# TODO Move away from needing nginx
# Setup Nginx.
ADD etc/nginx/nginx.conf /etc/nginx/
