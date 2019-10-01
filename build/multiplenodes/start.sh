#!/bin/sh
set -euf -o pipefail
# start statechain
start() {
  ssd start
}

start
