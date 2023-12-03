#!/bin/sh

SCRIPT_PATH=$(dirname $(realpath $0))
TAG=$1
if [ -z $TAG ]; then
  TAG='multi-life:latest'
fi
"$SCRIPT_PATH"/build.sh "-t $TAG" &&
docker run -it --rm -p 8080:80 -v "$SCRIPT_PATH"/assets:/assets $TAG
