SHELL = /bin/bash

TARGET       = apiserver
CLI_TARGET   = apicli
PROJECT_NAME = github.com/zanecloud/apiserver

MAJOR_VERSION = $(shell cat VERSION)
GIT_VERSION   = $(shell git log -1 --pretty=format:%h)
GIT_NOTES     = $(shell git log -1 --oneline)


IMAGE_NAME     = registry.cn-hangzhou.aliyuncs.com/zanecloud/apiserver
BUILD_IMAGE     = golang:1.8.3-onbuild

install:
	brew install mongodb redis npm

init:
	bash scripts/sbin/init.sh

apiserver:clean
	CGO_ENABLED=0  go build -a -installsuffix cgo -v -ldflags "-X ${PROJECT_NAME}/pkg/logging.ProjectName=${PROJECT_NAME}" -o ${TARGET}

autodeploy:clean-deploy
	cd tools && CGO_ENABLED=0  go build -a -installsuffix cgo -v -ldflags "-X ${PROJECT_NAME}/pkg/logging.ProjectName=${PROJECT_NAME}" -o autodeploy && cd ..

portal:
	cd static && npm install && npm run build && cd ..

build:
	docker run --rm -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE} make apiserver
	docker run --rm -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE} make autodeploy
image: build
	docker build --rm -t ${IMAGE_NAME}:${MAJOR_VERSION}-${GIT_VERSION} .
	docker tag  ${IMAGE_NAME}:${MAJOR_VERSION}-${GIT_VERSION} ${IMAGE_NAME}:${MAJOR_VERSION}
push: image
	docker push ${IMAGE_NAME}:${MAJOR_VERSION}-${GIT_VERSION}
	docker push ${IMAGE_NAME}:${MAJOR_VERSION}

shell:
	docker build --rm -t ${BUILD_IMAGE} contrib/builder/binary
	docker run -ti --rm -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE} /bin/bash

run:apiserver
	MONGO_URLS=127.0.0.1 MONGO_DB=zanecloud  ROOT_DIR=./static ./apiserver -l debug start

release:portal build
	rm -rf release && mkdir -p release/apiserver/bin
	cp -r static/public     release/apiserver/
	cp -r static/dist       release/apiserver/
	cp -r scripts/sbin     release/apiserver/
	cp -r scripts/systemd     release/apiserver/
	cp static/index.html release/apiserver/
	cp apiserver release/apiserver/bin/
	cp tools/autodeploy release/apiserver/bin/
	cd release && tar zcvf apiserver-${MAJOR_VERSION}-${GIT_VERSION}.tar.gz apiserver && cd ..


publish:release
	ssh -q root@${TARGET_HOST}  "mkdir -p /opt/zanecloud"
	scp release/apiserver-${MAJOR_VERSION}-${GIT_VERSION}.tar.gz  root@${TARGET_HOST}:/opt/zanecloud
	ssh -q root@${TARGET_HOST}  "cd /opt/zanecloud && rm -rf apiserver && tar zxvf apiserver-${MAJOR_VERSION}-${GIT_VERSION}.tar.gz"
	ssh -q root@${TARGET_HOST}  "systemctl stop apiserver"
	ssh -q root@${TARGET_HOST}  "systemctl start apiserver"

clean:
	rm -rf apiserver

clean-deploy:
	rm -rf autodeploy


test:
	mongo zanecloud --eval "db.user.remove({'name':'sadan'})"
	mongo zanecloud --eval "db.team.remove({})"
	mongo zanecloud --eval "db.pool.remove({})"
	mongo zanecloud --eval "db.node.remove({})"
	mongo zanecloud --eval "db.container.remove({})"
	cd handlers && go test -v

all:init portal run

.PHONY: image build local
