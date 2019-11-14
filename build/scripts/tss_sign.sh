#!/bin/sh

# wait until our config file is updated with a node id. This is done by the
# tss_keygen binary
while [ "$(cat $CONFIGFILE | jq .nodeid)" != "waiting" ]; do
  sleep 1
done

exec "$@"
