## ZANECLOUD - APISERVER

### APISERVER 简述

apiserver是用户访问zanecloud的入口，对pool进行管理和路由。

# 自动化测试

目前的自动化测试仅支持mac


在apiserver跟目录下执行


···
make run
···

则自动安装依赖的包，并启动mongodb/apiserver


在另一个terminal中运行
···
make test POOL_NAME=pool123
···
则将本机的docker注册为一个pool，名字在参数中指定；并在apiserver中为该pool起一个proxy，通过该proxy暴露docker/swarm接口服务。


