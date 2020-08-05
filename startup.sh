#!/bin/sh

echo "Starting sample Prometheus grep..."
while [ true ]; do cat test.txt; done | promgrep

