#!/usr/bin/env bash

USER_NAME=$1
USER_PASS=$2


echo "create user:$1"

UserID=`curl -sSL  -X POST -H "Content-Type: application/json"  -d @scripts/create-user.json "http://localhost:8080/users/create?Name=${USER_NAME}&Pass=${USER_PASS}" | jq .Id | tr -d "\""`


curl -sSL  -X POST -H "Content-Type: application/json"  http://localhost:8080/users/${UserID}/inspect

curl -sSL  -X POST -H "Content-Type: application/json"  http://localhost:8080/users/ps

curl -sSL  -X POST -H "Content-Type: application/json"  http://localhost:8080/users/${USER_NAME}/login?Pass=${USER_PASS}
