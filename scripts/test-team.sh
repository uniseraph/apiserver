#!/usr/bin/env bash

TEAM_NAME=$1


echo "create team:$1"
curl  -X POST -H "Content-Type: application/json"  -d @scripts/create-team.json "http://localhost:8080/teams/create?name=${TEAM_NAME}"



curl   http://localhost:8080/teams/${TEAM_NAME}/inspect
