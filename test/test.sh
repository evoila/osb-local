#!/bin/bash

set -eu

dir=$(dirname $0)

if [[ -z "$1" ]]; then
  exit 1
fi

SPACE=$1

if [[ -z "$2" ]]; then
  exit 1
fi

SERVICE=$2

if [[ -z "$3" ]]; then
  exit 1
fi

SERVICE_PLAN=$3

if [[ -z "$4" ]]; then
  exit 1
fi

SERVICE_INSTANCE=$4

if [[ -z "$5" ]]; then
  exit 1
fi

APP=$5

if [[ -z "$6" ]]; then
  exit 1
fi

APP_PATH=$6


echo "Starting test deployment"
pushd $dir > /dev/null

cf api https://api.dev.cfdev.sh  --skip-ssl-validation
cf auth admin admin

cf create-space -o system $SPACE
cf t -o system -s $SPACE

cf create-service $SERVICE $SERVICE_PLAN $SERVICE_INSTANCE

bosh int ./app.yml \
  -v hostname=$APP \
  -v service-instance=$SERVICE_INSTANCE \
  -v path=$APP_PATH \
  > ./manifest.yml

cf push $APP

curl "http://$APP.dev.cfdev.sh"

popd > /dev/null

echo "Finished test successfully"
exit 0