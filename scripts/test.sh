#!/usr/bin/env bash


POOL_NAME=$1

curl -i -X POST -H "Content-Type: application/json"  -d @scripts/create-pool.json http://localhost:8080/pools/register?name=${POOL_MAME}


curl -i -X GET http://localhost:8080/pools/${POOL_NAME}/inspect