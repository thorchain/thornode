#!/bin/sh

#
# Smoke Tests.
#
# Run our smoke tests against a Thorchain instance.
#

# Set the current working directory.
cd "$(dirname "$0")"

# Move into our Docker directory.
cd ../docker

make NET="$1" smoke-test
