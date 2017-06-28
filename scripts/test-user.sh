#!/usr/bin/env bash

USER_NAME=$1
USER_PASS=$2


echo "register user:$1"
curl  -i -X POST -H "Content-Type: application/json"  -d @scripts/create-user.json "http://localhost:8080/users/create?name=${USER_NAME}&pass=${USER_PASS}"



curl -i -X POST -H "Content-Type: application/json"  http://localhost:8080/users/${USER_NAME}/login?pass=${USER_PASS}
