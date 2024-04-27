#!/bin/bash
set -eu
go build
pkill ucms || true
sleep 2
# ./ucms
nohup ./ucms > ucms.log 2>&1 &
