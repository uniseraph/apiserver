## ZANECLOUD - APISERVER

### APISERVER 简述

apiserver是用户访问zanecloud的入口，进行用户权限管理，对pool进行管理和路由。



## 安装相关依赖包

下载代码到本机的 $GOPATH/src/github.com/zanecloud目录下


```
    make install
```

该命令只需要执行一次


## 重置测试环境
```
   make init
```


## 接口自动化测试

目前的自动化测试仅支持mac


在apiserver跟目录下执行

```
make run
```

则自动安装依赖的包，并启动mongodb/apiserver


在另一个terminal中运行
```
make test
```


## 运行完整环境

```
make install && make all
```
然后访问 http://localhost:8080

