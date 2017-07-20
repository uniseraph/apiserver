#!/usr/bin/env bash


BASE_DIR=$(cd `dirname $0` && pwd -P)


cp  ${BASE_DIR}/systemd/apiserver.service /etc/systemd/system/
mkdir -p $BASE_DIR}/etc/zanecloud && cp  systemd/apiserver /etc/zanecloud/apiserver

bash ${BASE_DIR}/sbin/init.sh
./bin/apiserver init