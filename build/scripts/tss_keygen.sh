#!/bin/sh

make -p ~/.tss/config

echo "{\"parties\":\"$PARTIES\",\"threshold\":\"$THRESHOLD\",\"bootstrapnode\":\"$BOOTSTRAPNODE\",\"signerserver\":\"$SIGNERSERVER\",\"keygenserver\":\"$KEYGENSERVER\",\"partynum\":\"$PARTYNUM\",\"nodeid\":\"$TSS_NODE_ID\"}" > $CONFIGFILE

exec "$@"
