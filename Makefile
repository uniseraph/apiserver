SHELL = /bin/bash

TARGET       = apiserver
CLI_TARGET   = apicli
PROJECT_NAME = github.com/zanecloud/apiserver

MAJOR_VERSION = $(shell cat VERSION)
GIT_VERSION   = $(shell git log -1 --pretty=format:%h)
GIT_NOTES     = $(shell git log -1 --oneline)


IMAGE_NAME     = registry.cn-hangzhou.aliyuncs.com/zanecloud/apiserver
BUILD_IMAGE     = golang:1.8

install:
	brew install mongodb redis npm

init:
	bash scripts/init.sh

apiserver:clean
	CGO_ENABLED=0  go build -a -installsuffix cgo -v -ldflags "-X ${PROJECT_NAME}/pkg/logging.ProjectName=${PROJECT_NAME}" -o ${TARGET}

apicli:cleancli
	CGO_ENABLED=0  go build -a -installsuffix cgo -v -ldflags "-X ${PROJECT_NAME}/pkg/logging.ProjectName=${PROJECT_NAME}"  -o ${CLI_TARGET}  client/client.go

portal:
	cd static && npm install && npm run build && cd ..

build:
	docker run --rm -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE} make apiserver apicli
image: build
	docker build --rm -t ${IMAGE_NAME}:${MAJOR_VERSION}-${GIT_VERSION} .
	docker tag  ${IMAGE_NAME}:${MAJOR_VERSION}-${GIT_VERSION} ${IMAGE_NAME}:${MAJOR_VERSION}
push: image
	docker push ${IMAGE_NAME}:${MAJOR_VERSION}-${GIT_VERSION}
	docker push ${IMAGE_NAME}:${MAJOR_VERSION}

shell:
	docker build --rm -t ${BUILD_IMAGE} contrib/builder/binary
	docker run -ti --rm -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE} /bin/bash

run:
	MONGO_URLS=127.0.0.1 MONGO_DB=zanecloud  ROOT_DIR=./static ./apiserver -l debug start

compose:
	docker-compose up -d

clean:
	rm -rf apiserver

cleancli:
	rm -rf apicli

test:apicli
	mongo zanecloud --eval "db.user.remove({'name':'sadan'})"
	mongo zanecloud --eval "db.team.remove({'name':'team1'})"
	./apicli

.PHONY: image build local
