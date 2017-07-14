#!/usr/bin/env bash


brew services restart mongodb redis

mongo zanecloud --eval "db.dropDatabase()"
mongo zanecloud --eval "db.user.createIndex({name:1}, {unique:true})"
mongo zanecloud --eval "db.team.createIndex({name:1}, {unique:true})"
mongo zanecloud --eval "db.pool.createIndex({name:1}, {unique:true})"
#TODO name + poolid做唯一性约束
mongo zanecloud --eval "db.application.createIndex({name:1,poolid:1}, {unique:true})"
mongo zanecloud --eval "db.container.createIndex({name:1,poolid:1}, {unique:true})"  #创建容器时候一开始不知道容器名字
mongo zanecloud --eval "db.container.createIndex({containerid:1,poolid:1}, {unique:true})"
mongo zanecloud --eval "db.container.createIndex({poolid:1})"
mongo zanecloud --eval "db.container.createIndex({applicationid:1})"


#准备加盐计算
name=root
salt="1234567891234567"
pass="hell05a"
content="$pass:$salt"
#生成加盐后的密码
encryptedPassword=$(md5 -qs $content)

mongo zanecloud --eval "db.user.insertOne({name:'$name',pass:'$encryptedPassword',salt: '$salt',roleset:4})"
