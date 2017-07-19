#!/usr/bin/env bash


cp  systemd/apiserver.service /etc/systemd/system/
mkdir -p /etc/zanecloud && cp  systemd/apiserver /etc/zanecloud/apiserver

bash sbin/init.sh
./bin/apiserver init