#!/usr/bin/env bash

POOL_NAME=$1

echo "create the pool"
curl  -X POST -H "Content-Type: application/json"  -d @scripts/create-pool.json http://localhost:8080/pools/register?name=${POOL_NAME}

docker rm -f test

DOCKER_HOST=`curl -sSL -X GET http://localhost:8080/pools/${POOL_NAME}/inspect | jq .ProxyEndpoints[0] | tr -d "\"" `


echo "DOCKER_HOST=${DOCKER_HOST}"

DOCKER_HOST=${DOCKER_HOST} docker pull nginx
DOCKER_HOST=${DOCKER_HOST} docker create --name test nginx
DOCKER_HOST=${DOCKER_HOST} docker start test
DOCKER_HOST=${DOCKER_HOST} docker logs test
DOCKER_HOST=${DOCKER_HOST} docker inspect test
DOCKER_HOST=${DOCKER_HOST} docker exec test pwd
DOCKER_HOST=${DOCKER_HOST} docker exec -ti test bash
DOCKER_HOST=${DOCKER_HOST} docker network ls
DOCKER_HOST=${DOCKER_HOST} docker images
DOCKER_HOST=${DOCKER_HOST} docker info
DOCKER_HOST=${DOCKER_HOST} docker version
DOCKER_HOST=${DOCKER_HOST} docker ps -a
DOCKER_HOST=${DOCKER_HOST} docker stop test
DOCKER_HOST=${DOCKER_HOST} docker rm test



