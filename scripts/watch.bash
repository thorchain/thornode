#!/bin/bash

make test ; fswatch . -e ".*" -i "\\.go$" | xargs -n1 -I{}  make clear test
