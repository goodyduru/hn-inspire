#!/bin/sh
set -o allexport
source $(dirname "$0")/.env
set +o allexport

# Update .env.example just in case
awk -F "=" '{print $1"="}' $(dirname "$0")/.env > $(dirname "$0")/.env.example