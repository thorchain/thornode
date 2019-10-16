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
  echo "Usage: $0 -r <rpc host> -g <target_group> -f <faucet key> -p <pool key> -e <environment>" 1>&2;
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

  if [ $HEIGHT < 500 ]; then
    return 0
  else
    return 1
  fi
}

# Check the supplied opts.
while getopts ":r:g:f:p:e:" o; do
    case "${o}" in
        r)
            r=${OPTARG}
            ;;
        g)
            g=${OPTARG}
            ;;
        f)
            f=${OPTARG}
            ;;
        p)
            p=${OPTARG}
            ;;
        e)
            e=${OPTARG}
            ;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

if [ -z "${r}" ] ||
    [ -z "${g}" ] ||
    [ -z "${f}" ] ||
    [ -z "${p}" ] ||
    [ -z "${e}" ]; then
    usage
fi

# Ensures we don't run forever!
COUNT=0
MAX_ATTEMPTS=30

# Target group ARN.
TG_ARN=$(aws elbv2 describe-target-groups | jq -r --arg TG "${g}" '.TargetGroups[] | select(.TargetGroupName==$TG)' | jq -r '.TargetGroupArn')

# Loop through our targets and check the health.
check_health $TG_ARN

while [ $? -ne 0 ]; do
  sleep 15

  COUNT="$((COUNT+1))"
  if [ $COUNT -eq $MAX_ATTEMPTS ]; then
    break;
  fi

  check_health $TG_ARN
done

# Run our smoke tests.
if [ $COUNT -lt $MAX_ATTEMPTS ]; then
  check_block_height "${n}"
  if [ $? -eq 0 ]; then
    # Smoke 'em if you got 'em.
    make FAUCET_KEY="${f}" POOL_KEY="${p}" ENV="${e}" -C ../../ smoke-test-refund
  else
    echo "Exiting....looks like this chain was started a while ago?"
    exit
  fi
else
  echo "Exiting...max attempts reached. Maybe increase the timeout?"
  exit
fi
