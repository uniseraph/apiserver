FROM alpine:3.4

MAINTAINER zhengtao.wuzt <zhengtao.wuzt@gmail.com>

ADD  ./apiserver  /usr/bin/apiserver
RUN chmod +x /usr/bin/apiserver


ENTRYPOINT ["/usr/bin/apiserver"]


