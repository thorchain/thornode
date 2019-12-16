#!/bin/sh

set -ex

# wait for our private key
while [ ! -f $PRIVKEY ]; do
    sleep 3
done

if [ ! -z ${SEED+x} ]; then
    while ! nc -z $SEED $SEEDHTTPPORT; do
        sleep 1
    done

    cat $PRIVKEY | /go/bin/tss -loglevel debug -http $TSSHTTPPORT -port $TSSP2PPORT -peer /ip4/$SEED/tcp/$SEEDP2PPORT/ipfs/$(curl http://$SEED:$SEEDHTTPPORT/p2pid)

else
    cat $PRIVKEY | /go/bin/tss -loglevel debug -http $TSSHTTPPORT -port $TSSP2PPORT
fi
