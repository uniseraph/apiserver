#!/usr/bin/env bash



mongo zanecloud --eval "db.dropDatabase()"
mongo zanecloud --eval "db.user.createIndex({name:1}, {unique:true})"
mongo zanecloud --eval "db.team.createIndex({name:1}, {unique:true})"
mongo zanecloud --eval "db.pool.createIndex({name:1}, {unique:true})"
#TODO name + poolid做唯一性约束
mongo zanecloud --eval "db.application.createIndex({name:1,poolid:1}, {unique:true})"
mongo zanecloud --eval "db.env_tree_node_param_key.createIndex({name:1,tree:1}, {unique:true})"
mongo zanecloud --eval "db.container_audit_trace.createIndex({token:1}, {unique:true})"
mongo zanecloud --eval "db.container.createIndex({name:1,poolid:1}, {unique:false})"
mongo zanecloud --eval "db.container.createIndex({containerid:1,poolid:1}, {unique:true})"
mongo zanecloud --eval "db.container.createIndex({poolid:1})"
mongo zanecloud --eval "db.container.createIndex({applicationid:1})"
mongo zanecloud --eval "db.container_audit_log.createIndex({operation:1})"
mongo zanecloud --eval "db.container_audit_log.createIndex({token:1})"
mongo zanecloud --eval "db.container_audit_trace.createIndex({token:1})"
