#!/bin/sh

set -ex

# wait for our private key
while [ ! -f $PRIVKEY ]; do
    sleep 3
done

if [ ! -z ${SEED+x} ]; then
    while ! nc -z $SEED $SEEDP2PPORT; do
        sleep 1
    done

    cat $PRIVKEY | /go/bin/tss -home ~/.tss -loglevel debug -info-port $INFOPORT -tss-port $TSSPORT -p2p-port $P2PPORT -peer /ip4/$SEED/tcp/$SEEDP2PPORT/ipfs/$(curl http://$SEED:$SEEDINFOPORT/p2pid)

else
    cat $PRIVKEY | /go/bin/tss -home ~/.tss -loglevel debug -info-port $INFOPORT -tss-port $TSSPORT -p2p-port $P2PPORT
fi
