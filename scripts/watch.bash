#!/bin/bash

make test ; fswatch . -e ".*" -i "\\.go$" --event PlatformSpecific | xargs -n1 -I{}  make clear test
