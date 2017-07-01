FROM alpine:3.4

MAINTAINER zhengtao.wuzt <zhengtao.wuzt@gmail.com>

ADD  ./apiserver  /usr/bin/apiserver
RUN chmod +x /usr/bin/apiserver

RUN mkdir -p /opt/zanecloud/portal
ADD ./index.html /opt/zanecloud/portal
ADD ./public /opt/zanecloud/portal
ADD ./dist   /opt/zanecloud/portal


ENV ROOT_DIR /opt/zanecloud/portal

ENTRYPOINT ["/usr/bin/apiserver"]


