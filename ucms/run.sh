#!/bin/bash
set -eu

file=GeoLite2-Country.mmdb
if [ -f "$file" ]; then
  echo "File '$file' exists."
else
  echo "File '$file' does not exist. Please download."
  exit 1
fi
go build
pkill ucms || true
sleep 2
nohup ./ucms > ucms.log 2>&1 &
