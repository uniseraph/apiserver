#!/usr/bin/env bash

./autodeploy start --template_uuid=59716d6aaad8c0484227b101 --application_uuid=597191ceaad8c0484227b11d --image_name=nginx \
    --image_tag=1.7 --user=root --pass=hell05a --service_name=webcenter --apiserver_host=localhost:8080