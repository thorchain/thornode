#
# Observer, Signer, Statechain
#

#
# Build
#
FROM golang:1.13 AS build

WORKDIR /go/src/app
RUN git config --global http.sslVerify "false"

COPY . .

RUN GO111MODULE=on go mod verify
RUN GO111MODULE=on go get -d -v ./...

WORKDIR /go/src/app/cmd/ssd
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ssd .

WORKDIR /go/src/app/cmd/sscli
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sscli .

WORKDIR /go/src/app/cmd/observed
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o observed .

WORKDIR /go/src/app/cmd/signd
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o signd .

#
# Main
#
FROM golang:1.13.3-alpine

RUN apk add --update linux-headers build-base jq supervisor git nginx && \
    rm -rf /var/cache/apk/*

ENV PATH="${PATH}:/go/bin"
ENV GOPATH=/go

# Copy the compiled binaires over.
COPY --from=build /go/src/app/cmd/ssd/ssd /go/bin/ssd
COPY --from=build /go/src/app/cmd/sscli/sscli /go/bin/sscli
COPY --from=build /go/src/app/cmd/observed/observed /go/bin/observed
COPY --from=build /go/src/app/cmd/signd/signd /go/bin/signd

# Add users.
RUN adduser -Ds /bin/sh supervisor
RUN adduser -Ds /bin/sh www-data -G www-data

# TODO Move away from needing supervisor
# Setup supervisor.
RUN mkdir -p /var/log/supervisor
RUN mkdir -p /var/run/supervisor
RUN mkdir -p /etc/supervisor/conf.d
ADD etc/supervisor.conf /etc/
ADD etc/supervisor/conf.d/*.ini /etc/supervisor/conf.d/

# TODO Move away from needing nginx
# Setup Nginx.
ADD etc/nginx/nginx.conf /etc/nginx/

# Setup Observer/Signer.
RUN mkdir -p /etc/observe/observed
RUN mkdir /etc/observe/signd

EXPOSE 9000
EXPOSE 1317
EXPOSE 81

CMD ["supervisord", "-c", "/etc/supervisor.conf"]
