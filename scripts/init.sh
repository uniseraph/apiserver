#!/usr/bin/env bash


brew services stop mongodb

rm -rf /usr/local/var/mongodb  && mkdir -p /usr/local/var/mongodb

brew services restart mongodb

mongo zanecloud --eval "db.user.createIndex({name:1}, {unique:true})"
#mongo zanecloud --eval "db.pool.createIndex({name:1}, {unique:true})"
#mongo zanecloud --eval "db.container.createIndex({name:1}, {unique:true})"
#mongo zanecloud --eval "db.container.createIndex({id:1}, {unique:true})"


brew services restart redis