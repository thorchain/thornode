#!/bin/sh

# wait for our config file to exist
while [ ! -f $CONFIGFILE ]; do
    sleep 3
done
# wait until our config file is updated with a node id. This is done by the
# tss_keygen binary

while [ "$(cat $CONFIGFILE | jq .nodeid)" != "waiting" ]; do
    sleep 3
done

exec "$@"
