#!/bin/bash

if [[ -z "$1" ]]; then
  exit 1
fi

SPACE=$1

if [[ -z "$2" ]]; then
  exit 1
fi

SERVICE_INSTANCE=$2

if [[ -z "$3" ]]; then
  exit 1
fi

APP=$3


cf api https://api.dev.cfdev.sh  --skip-ssl-validation
cf auth admin admin

cf create-space -o system $SPACE
cf t -o system -s $SPACE

cf delete music -r -f
cf delete-service $SERVICE_INSTANCE -f

exit 0