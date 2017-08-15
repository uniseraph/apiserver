# 安装文档


## 第一次安装

### 解压apiserver安装文件到/opt/zanelcoud目录下
```
mkdir -p /opt/zanecloud  && cd /opt/zanecloud  && tar zxvf apiserver-1.0.1-xxxxx.tar.gz
```
xxxxx是gitcommit，请参考具体apiserver安装文件名


### 安装

```
cd /opt/zanecloud/apiserver && bash -x sbin/install.sh
```


### 检查安装是否成功

```
systemctl status apiserver

```

## 升级发布
### 解压apiserver安装文件到/opt/zanecloud目录下，

```
rm -rf /opt/zanecloud/apiserver   && tar zxvf apiserver-1.0.1-xxxxx.tar.gz -C /opt/zanecloud
```

### 如果配置发生变化则修改apiserver 配置文件

```
vi /etc/zanecloud/apiserver.conf
```

### 重启apiserver服务
```
systemctl restart apiserver
```

