#!/bin/sh

TAG=$1
if [ -z $TAG ]; then
  TAG='multi-life:latest'
fi
export DOCKER_BUILDKIT=1
docker build . --target js-lint &&
docker build . --target go-lint &&
docker build . --target go-test &&
docker build . --target bin -t $TAG
