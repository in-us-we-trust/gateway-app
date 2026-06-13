#!/bin/bash

set -a
source .env
set +a

PORT=8090 go run scripts.go run