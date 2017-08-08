#!/usr/bin/env bash

MASTER_HOST="172.16.99.148"
SLAVE_HOST="172.16.99.149"
ARBITER_HOST="172.16.99.150"
RSSET="zanecloud-rs"


cat <<-EOF
Mongo Replica Set will deploy to below Machines:

MASTER_HOST:    $MASTER_HOST
SLAVE_HOST:     $SLAVE_HOST
ARBITER_HOST:   $ARBITER_HOST

EOF

read -p "press y to continue, or stop and change hosts.[y/n]" -n 1 confirm

if [ $confirm -ne "y" ]; then
    exit -1
fi

HOSTS=(
$MASTER_HOST
$SLAVE_HOST
$ARBITER_HOST
)

for host in ${HOSTS[*]}; do

echo "INSTALL MONGO at ${host}"

cat <<-EOF | ssh -q root@${host}
if type apt-get >/dev/null 2>&1; then
  echo 'using apt-get '
  sudo apt-get update && apt-get install -y mongodb

elif type yum >/dev/nul 2>&1; then
  echo 'using yum'
  sudo yum install -y mongodb

else
  echo "no apt-get and no yum, exit"
  exit
fi

#
# after install
# config file is
#
#root@iZbp1fyry0dgojwdhct65xZ:~# cat /etc/mongodb.conf | grep -v "^#" | grep -v "^$"
#

echo > /etc/mongodb.conf
echo dbpath=/var/lib/mongodb >> /etc/mongodb.conf
echo logpath=/var/log/mongodb/mongodb.log >> /etc/mongodb.conf
echo logappend=true >> /etc/mongodb.conf
echo bind_ip=0.0.0.0 >> /etc/mongodb.conf
echo journal=true >> /etc/mongodb.conf
echo replSet=${RSSET} >> /etc/mongodb.conf

echo "RESTART MONGO at ${host}"

service mongodb restart

EOF

done;

echo "CONFIG MONGO RS at ${MASTER_HOST}"

CONFIG="{ _id:'${RSSET}', members:[ {_id:0,host:'${MASTER_HOST}:27017',priority:2}, {_id:1,host:'${SLAVE_HOST}:27017',priority:1}, {_id:2,host:'${ARBITER_HOST}:27017',arbiterOnly:true}] }"

cat <<-EOF | ssh -q root@${MASTER_HOST}
mongo zanecloud --eval "rs.initiate(${CONFIG})"
mongo zanecloud --eval "rs.conf()"
EOF

echo "Wait for replica set config."
sleep 10

cat <<-EOF | echo
DATA Migrate!!!
Below command will migrate data from single mongo to replica mongo.

    mongodump -h localhost -d zanecloud
    mongorestore -h ${MASTER_HOST}:27017 -d zanecloud  --directoryperdb ./dump/zanecloud/
EOF