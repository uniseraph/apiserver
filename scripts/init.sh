#!/usr/bin/env bash


brew services stop mongodb

rm -rf /usr/local/var/mongodb  && mkdir -p /usr/local/var/mongodb

brew services restart mongodb

mongo zanecloud --eval "db.user.createIndex({Name:1}, {unique:true})"
#mongo zanecloud --eval "db.pool.createIndex({Name:1}, {unique:true})"
#mongo zanecloud --eval "db.container.createIndex({Name:1}, {unique:true})"
#mongo zanecloud --eval "db.container.createIndex({Id:1}, {unique:true})"


brew services restart redis