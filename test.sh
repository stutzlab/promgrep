#!/bin/sh

set -e
# set -x

echo "Starting sample Prometheus grep..."
while [ true ]; do 
    cat /app/test.txt; sleep 0.1; done | /bin/promgrep --summary "exiting@EXITING NOW" --summary "test1@was ([0-9]+)ms"


