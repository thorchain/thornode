#!/bin/sh

[ $(gofumpt -l . | wc -l) -eq 0 ] && true
