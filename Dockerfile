FROM alpine:3.4
#FROM ubuntu:14.04
#FROM acs-reg.alipay.com/acs/alpine-base:1.0.0-ea1b016-20160901
MAINTAINER zhengtao.wuzt <zhengtao.wuzt@gmail.com>

RUN mkdir -p /opt/acs
ADD  ./apiserver  /usr/bin/apiserver
RUN chmod +x /usr/bin/apiserver


ENTRYPOINT ["/usr/bin/apiserver"]


