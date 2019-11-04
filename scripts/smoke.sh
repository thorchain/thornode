#!/bin/sh

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
  echo "Usage: $0 -r <rpc host> -g <target_group> -c <cluster> -s <service> -n <task count> -f <faucet key> -p <pool key> -e <environment>" 1>&2;
  exit 1;
}

#
# Task Count
#
task_count() {
  TASK_COUNT=$(aws ecs describe-services --cluster "${1}" --service "${2}" | jq '.services[0].deployments' | jq length)

  if [ $TASK_COUNT -gt $3 ]; then
    echo "New task(s) being provisioned. Waiting...."
    return 1
  else
    return 0
  fi
}

#
# Target Group Health
#
target_group_health() {
  TG_LENGTH=$(aws elbv2 describe-target-health --target-group-arn "${1}" | jq '.TargetHealthDescriptions' | jq length)
  END="$((TG_LENGTH-1))"

  for i in $(seq $END 0); do
    HEALTH=$(aws elbv2 describe-target-health --target-group-arn "${1}" | jq -r ".TargetHealthDescriptions[$i].TargetHealth.State")
    if [ $HEALTH != 'healthy' ]; then
      echo "Unhealthy node detected. Waiting...."
      return 1
    fi
  done
}

#
# Wrapper for task_count()
#
check_tasks() {
  task_count "${1}" "${2}" $3
}

#
# Wrapper for target_group_health()
#
check_health() {
  target_group_health "${1}"
}

#
# Check the block height.
#
check_block_height() {
  HEIGHT=$(curl -s "$1/block" | jq -r '.result.block_meta.header.height')

  if [ $HEIGHT -lt 500 ]; then
    return 0
  else
    return 1
  fi
}

# Check the supplied opts.
while getopts ":r:g:c:s:n:f:p:e:" o; do
    case "${o}" in
        r)
            r=${OPTARG}
            ;;
        g)
            g=${OPTARG}
            ;;
        c)
            c=${OPTARG}
            ;;
        s)
            s=${OPTARG}
            ;;
        n)
            n=${OPTARG}
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

# All opts provided?
if [ -z "${r}" ] ||
    [ -z "${g}" ] ||
    [ -z "${c}" ] ||
    [ -z "${s}" ] ||
    [ -z "${n}" ] ||
    [ -z "${f}" ] ||
    [ -z "${p}" ] ||
    [ -z "${e}" ]; then
  usage
fi

# AWS ENV vars set?
if [ -z "$AWS_ACCESS_KEY_ID" ] ||
    [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
  echo "AWS ENV's not set!"
  exit 1
fi

# Ensures we don't run forever!
COUNT=0
MAX_ATTEMPTS=30

# Check the number of tasks - this tells us if a new task is in the process of being provisioned..
check_tasks "${c}" "${s}" "${n}"

while [ $? -ne 0 ]; do
  sleep 15

  COUNT="$((COUNT+1))"
  if [ $COUNT -eq $MAX_ATTEMPTS ]; then
    break;
  fi

  check_tasks "${c}" "${s}" ${n}
done

# This would happen if the task count supplied did not match
# (e.g: there are always two running tasks but we supplied a
# task count of 1 to the script.
if [ $COUNT -eq $MAX_ATTEMPTS ]; then
  echo "Exiting. Either the supplied task count is wrong, or the new task(s) are not booting correctly."
  exit 1
else
  # Reset the counter.
  COUNT=0
fi

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
  check_block_height "${r}"
  if [ $? -eq 0 ]; then
    # Smoke 'em if you got 'em.
    echo "Running: smoke-test-audit...."
    make FAUCET_KEY="${f}" POOL_KEY="${p}" ENV="${e}" -C ../../ smoke-test-audit

    echo "Running: smoke-test-refund...."
    make FAUCET_KEY="${f}" POOL_KEY="${p}" ENV="${e}" -C ../../ smoke-test-refund
  else
    echo "Exiting. Looks like this chain was started a while ago?"
    exit 1
  fi
else
  echo "Exiting. Max attempts reached. Maybe increase the timeout?"
  exit 1
fi
