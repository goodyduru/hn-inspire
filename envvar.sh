#!/bin/sh
export $(cat $(dirname "$0")/.env | xargs) && env

# Update .env.example just in case
awk -F "=" '{print $1"="}' $(dirname "$0")/.env > $(dirname "$0")/.env.example