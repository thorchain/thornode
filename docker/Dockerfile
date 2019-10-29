#
# Observer, Signer, Statechain
#

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
RUN go build -a -installsuffix cgo -o ssd ./cmd/ssd
RUN go build -a -installsuffix cgo -o sscli ./cmd/sscli
RUN go build -a -installsuffix cgo -o observed ./cmd/observed
RUN go build -a -installsuffix cgo -o signd ./cmd/signd

#
# Main
#
FROM alpine

RUN apk add --update jq curl supervisor nginx && \
    rm -rf /var/cache/apk/*

ENV PATH="${PATH}:/go/bin"

# Copy the compiled binaires over.
COPY --from=build /go/src/app/generate /go/bin/
COPY --from=build /go/src/app/ssd /go/bin/
COPY --from=build /go/src/app/sscli /go/bin/
COPY --from=build /go/src/app/observed /go/bin/
COPY --from=build /go/src/app/signd /go/bin/

#ENV PASTEBINIT_SHA256="8f76d6b194c3063bf4ebbe4e2761d13f79da92106796a59e28f599f9570eac6c"
#RUN curl -fSL "https://github.com/jessfraz/pastebinit/releases/download/v0.2.4/pastebinit-linux-amd64" -o "/usr/local/bin/pastebinit" \
	#&& echo "${PASTEBINIT_SHA256}  /usr/local/bin/pastebinit" | sha256sum -c - \
	#&& chmod a+x "/usr/local/bin/pastebinit"

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

EXPOSE 9000
EXPOSE 1317
EXPOSE 81

CMD ["supervisord", "-c", "/etc/supervisor.conf"]
