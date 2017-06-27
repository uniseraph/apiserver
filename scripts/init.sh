#!/usr/bin/env bash


mongo zanecloud --eval "db.dropDatabase()"
mongo zanecloud --eval "db.user.createIndex({name:1}, {unique:true})"
mongo zanecloud --eval "db.team.createIndex({name:1}, {unique:true})"
mongo zanecloud --eval "db.pool.createIndex({name:1}, {unique:true})"
mongo zanecloud --eval "db.container.createIndex({name:1}, {unique:true})"
mongo zanecloud --eval "db.container.createIndex({id:1}, {unique:true})"


brew services restart redis