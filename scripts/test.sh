#!/usr/bin/env bash


curl -i -X POST -H "Content-Type: application/json"  -d @scripts/create-pool.json http://localhost:8080/pools/register?name=pool106


curl -i -X GET http://localhost:8080/pools/pool1/inspect