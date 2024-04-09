#!/bin/bash
set -eu

go build
nohup ./ucms > ucms.log 2>&1 &
