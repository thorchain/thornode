
FROM thornode

# ENV GO111MODULE=on

COPY ./scripts/entrypoint.sh /entrypoint.sh
COPY ./tools/generate/generate.go /go/generate.go
# ENTRYPOINT /entrypoint.sh

