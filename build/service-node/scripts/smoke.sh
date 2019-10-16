#!/usr/bin/env bash

#
# Smoke Tests.
#
# This script will check to see if a Statechain was recently started,
# and if so, run our set of smoke tests against it.
#

#
# Usage
#
usage() {
  echo "Usage: $0 -n <node> -g <target_group>" 1>&2;
  exit 1;
}

#
# Target Group Size
#
target_group_health() {
  TG_LENGTH=$(aws elbv2 describe-target-health --target-group-arn "${1}" | jq '.TargetHealthDescriptions' | jq length)
  END="$((TG_LENGTH-1))"

  for i in $(seq $END 0); do
    HEALTH=$(aws elbv2 describe-target-health --target-group-arn "${1}" | jq -r ".TargetHealthDescriptions[$i].TargetHealth.State")
    if [ $HEALTH != 'healthy' ]; then
      echo "Unhealthy node detected!"
      return 1
    fi
  done
}

#
# Wrapper for target_group_health()
#
check_health() {
  echo "Checking target group health...."
  target_group_health $1
}

#
# Check the block height.
#
check_block_height() {
  HEIGHT=$(curl -s "$1/block" | jq -r '.result.block_meta.header.height')

  if [ $HEIGHT > 200 ]; then
    return 1
  fi
}

# Check the supplied opts.
while getopts ":n:g:" o; do
    case "${o}" in
        n)
            n=${OPTARG}
            ;;
        g)
            g=${OPTARG}
            ;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

if [ -z "${n}" ] || [ -z "${g}" ]; then
    usage
fi

# Ensures we don't run forever!
COUNT=0
MAX_ATTEMPTS=30

# Loop through our targets and check the health.
check_health "${g}"

while [ $? -ne 0 ]; do
  sleep 15

  COUNT="$((COUNT+1))"
  if [ $COUNT -eq $MAX_ATTEMPTS ]; then
    break;
  fi

  check_health "${g}"
done

# Get the block height.
if [ $COUNT -lt $MAX_ATTEMPTS ]; then
  check_block_height "${n}"
  if [ $? -eq 0 ]; then
    # Smoke 'em if you got 'em.
    echo "Run our tests!"
  else
    echo "Exiting. Looks like this chain was started a while ago?"
  fi
fi
