#!/usr/bin/env bash



if type apt-get >/dev/null 2>&1; then
  echo 'using apt-get '
  sudo apt-get update && apt-get install -y mongodb mongodb-server redis-server
  sudo systemctl restart mongodb
  sudo systemctl enable mongodb

elif type yum >/dev/nul 2>&1; then
  echo 'using yum'
  sudo yum install -y redis  mongodb mongodb-server
  sudo systemctl restart mongod
  sudo systemctl enable mongod

else
  echo "no apt-get and no yum, exit"
  exit
fi

systemctl restart redis
systemctl enable redis



BASE_DIR=$(cd `dirname $0` && cd .. && pwd -P)


cp  ${BASE_DIR}/systemd/apiserver.service /etc/systemd/system/
mkdir -p /etc/zanecloud && cp -f  systemd/apiserver.conf /etc/zanecloud/apiserver.conf

bash ${BASE_DIR}/sbin/init.sh
./bin/apiserver init



systemctl restart apiserver
systemctl enable apiserver
