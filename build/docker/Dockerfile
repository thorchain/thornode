#
# Observer, Signer, Statechain
#

#
# Build
#
FROM golang:1.13 AS build

WORKDIR /app

COPY . .

ENV GOBIN=/go/bin
ENV GOPATH=/go
ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go mod verify
RUN go get -d -v ./...

RUN go build -a -installsuffix cgo -o $GOBIN/thord ./cmd/thord
RUN go build -a -installsuffix cgo -o $GOBIN/thorcli ./cmd/thorcli
RUN go build -a -installsuffix cgo -o $GOBIN/observed ./cmd/observed
RUN go build -a -installsuffix cgo -o $GOBIN/signd ./cmd/signd
RUN go build -a -installsuffix cgo -o $GOBIN/generate ./tools/generate

# install keygenClient
RUN git clone https://gitlab.com/thorchain/tss/tss-tools.git /tmp/tss-tools
WORKDIR /tmp/tss-tools/cmd/keygenclient
RUN go mod verify
RUN go get -d -v ./...
RUN go build -a -installsuffix cgo -o $GOBIN/keygenclient .

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
COPY --from=build /go/bin/observed /usr/bin/
COPY --from=build /go/bin/signd /usr/bin/
COPY --from=build /go/bin/keygenclient /usr/bin/

# Add users.
RUN adduser -Ds /bin/sh www-data -G www-data

# TODO Move away from needing nginx
# Setup Nginx.
ADD etc/nginx/nginx.conf /etc/nginx/

EXPOSE 9000
EXPOSE 1317
EXPOSE 81
