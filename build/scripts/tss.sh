#!/bin/sh

set -ex

# wait for our private key
while [ ! -f $PRIVKEY ]; do
    sleep 3
done

if [ ! -z ${SEED+x} ]; then
    while ! nc -z $SEED 4040; do
        sleep 1
    done

    cat $PRIVKEY | /go/bin/tss -http 4040 -peer /ip4/$SEED/tcp/5040/ipfs/$(curl http://$SEED:4040/p2pid) -port 5040

else
    cat $PRIVKEY | /go/bin/tss -http 4040 -port 5040
fi
