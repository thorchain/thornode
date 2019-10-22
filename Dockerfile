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
FROM golang:alpine

ARG chain_id
ARG pool_address
ARG dex_host
ARG db_path
ARG rpc_host
ARG chain_host
ARG signer_name
ARG signer_passwd
ARG binance_private_key
ARG binance_test_net
ARG trusted_bnb_address
ARG start_block_height

ENV CHAIN_ID=$chain_id
ENV POOL_ADDRESS=$pool_address
ENV DEX_HOST=$dex_host
ENV DB_PATH=$db_path
ENV RPC_HOST=$rpc_host
ENV CHAIN_HOST=$chain_host
ENV SIGNER_NAME=$signer_name
ENV SIGNER_PASSWD=$signer_passwd
ENV BINANCE_PRIVATE_KEY=$binance_private_key
ENV BINANCE_TESTNET=$binance_test_net
ENV TRUSTED_BNB_ADDRESS=$trusted_bnb_address
ENV START_BLOCK_HEIGHT=$start_block_height

RUN apk add --update jq supervisor nginx && \
    rm -rf /var/cache/apk/*

ENV PATH="${PATH}:/go/bin"

# Copy the compiled binaires over.
COPY --from=build /go/src/app/cmd/ssd/ssd /go/bin/ssd
COPY --from=build /go/src/app/cmd/sscli/sscli /go/bin/sscli
COPY --from=build /go/src/app/cmd/observed/observed /go/bin/observed
COPY --from=build /go/src/app/cmd/signd/signd /go/bin/signd

RUN mkdir -p $DB_PATH/observer
RUN mkdir $DB_PATH/signer

# Add users.
RUN adduser -Ds /bin/bash supervisor
RUN adduser -Ds /bin/bash www-data -G www-data

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

# Generate a new key.
# TODO add back in.
RUN echo $SIGNER_PASSWD | sscli keys add statechain

# Setup Observer/Signer.
RUN mkdir -p /etc/observe/observed
RUN mkdir /etc/observe/signd
ADD etc/config.json /go/src/app/etc/config.json
RUN cat /go/src/app/etc/config.json | jq \
  --arg START_BLOCK_HEIGHT "$START_BLOCK_HEIGHT" \
  --arg CHAIN_ID "$CHAIN_ID" \
  --arg POOL_ADDRESS "$POOL_ADDRESS" \
  --arg RUNE_ADDRESS "$(sscli keys show statechain -a)" \
  --arg RPC_HOST "$RPC_HOST" \
  --arg OBSERVER_DB_PATH "$DB_PATH/observer" \
  --arg SIGNER_DB_PATH "$DB_PATH/signer" \
  --arg CHAIN_ID "$CHAIN_ID" \
  --arg CHAIN_HOST "$CHAIN_HOST" \
  --arg SIGNER_NAME "$SIGNER_NAME" \
  --arg SIGNER_PASSWD "$SIGNER_PASSWD" \
  --arg PRIVATE_KEY "$BINANCE_PRIVATE_KEY" \
  --arg DEX_HOST "$DEX_HOST" \
  '.chain_id = $CHAIN_ID | \
  .pool_address = $POOL_ADDRESS | \
  .rune_address = $RUNE_ADDRESS | \
  .dex_host = $DEX_HOST | \
  .observer_db_path = $OBSERVER_DB_PATH | \
  .signer_db_path = $SIGNER_DB_PATH | \
  .block_scanner["rpc_host"] = $RPC_HOST | \
  .block_scanner["start_block_height"] = $START_BLOCK_HEIGHT | \
  .state_chain["chain_id"] = $CHAIN_ID | \
  .state_chain["chain_host"] = $CHAIN_HOST | \
  .state_chain["signer_name"] = $SIGNER_NAME | \
  .state_chain["signer_passwd"] = $SIGNER_PASSWD | \
  .binance["private_key"] = $PRIVATE_KEY | \
  .binance["dex_host"] = $DEX_HOST' > /etc/observe/observed/config.json
RUN cat /etc/observe/observed/config.json | jq \
  '.block_scanner["start_block_height"] = 1 | \
  .block_scanner["rpc_host"] = "127.0.0.1:26657"' > /etc/observe/signd/config.json

# Setup statechain.
RUN ssd init local --chain-id statechain
RUN ssd add-genesis-account $(sscli keys show statechain -a) 100000000000thor
RUN sscli config chain-id statechain
RUN sscli config output json
RUN sscli config indent true
RUN sscli config trust-node true
RUN cat ~/.ssd/config/genesis.json | jq --arg POOL_ADDRESS "$POOL_ADDRESS" --arg NODE_ADDRESS "$(sscli keys show statechain -a)" --arg OBSERVER_ADDRESS "$(sscli keys show statechain -a)" --arg VALIDATOR "$(ssd tendermint show-validator)" '.app_state.swapservice.node_accounts[0] = {"node_address": $NODE_ADDRESS ,"status":"active","bond_address":$POOL_ADDRESS,"accounts":{"bnb_signer_acc":$POOL_ADDRESS, "bepv_validator_acc": $VALIDATOR, "bep_observer_acc": $OBSERVER_ADDRESS}}' > /go/src/app/genesis.json
RUN mv /go/src/app/genesis.json ~/.ssd/config/genesis.json
RUN cat ~/.ssd/config/genesis.json
RUN ssd validate-genesis

EXPOSE 9000
EXPOSE 1317
EXPOSE 81

CMD ["supervisord", "-c", "/etc/supervisor.conf"]
