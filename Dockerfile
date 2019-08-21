#
# BEPSwap Statechain
#
FROM golang:1.12

RUN apt-get update
RUN apt-get install -y jq supervisor

# Setup Supervisor
RUN useradd -ms /bin/bash supervisor
RUN mkdir -p /var/log/supervisor
RUN mkdir -p /var/run/supervisor
RUN chown supervisor:supervisor /var/log/supervisor
RUN chown supervisor:supervisor /var/run/supervisor
RUN mkdir -p /etc/supervisor/conf.d
ADD supervisor.conf /etc/supervisor.conf

WORKDIR /go/src/app
RUN git config --global http.sslVerify "false"
RUN git clone https://gitlab.com/thorchain/bepswap/statechain.git

# Setup Statechain
WORKDIR /go/src/app/statechain
RUN go mod download
RUN make setup
EXPOSE 1317

CMD ["supervisord", "-c", "/etc/supervisor.conf"]
