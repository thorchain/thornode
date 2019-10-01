FROM golang:1.13 as build

WORKDIR /go/src/app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY ./ ./
WORKDIR /go/src/app/cmd/ssd
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ssd .

WORKDIR /go/src/app/cmd/sscli
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sscli .

FROM alpine:latest
ARG node_id
ENV NODE_ID=$node_id
ENV SS_HOME="/statechain/$NODE_ID/ssd"
ENV SSC_HOME="/statechain/$NODE_ID/sscli/"
VOLUME ["/statechain"]

RUN apk add --update jq curl

WORKDIR /usr/bin
COPY --from=build /go/src/app/cmd/ssd/ssd /usr/bin/ssd
COPY --from=build /go/src/app/cmd/sscli/sscli /usr/bin/sscli

COPY build/multiplenodes/init.sh /usr/bin/init.sh
RUN chmod +x /usr/bin/init.sh

COPY build/multiplenodes/second.sh /usr/bin/second.sh
RUN chmod +x /usr/bin/second.sh

COPY build/multiplenodes/start.sh /usr/bin/start.sh
RUN chmod +x /usr/bin/start.sh

CMD ["/usr/bin/start.sh"]