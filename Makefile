SHELL = /bin/bash

TARGET       = apisevevr
PROJECT_NAME = github.com/zanecloud/apiserver

MAJOR_VERSION = $(shell cat VERSION)
GIT_VERSION   = $(shell git log -1 --pretty=format:%h)
GIT_NOTES     = $(shell git log -1 --oneline)


IMAGE_NAME     = github.com/zanecloud/apiserver
BUILD_IMAGE     = golang:1.8


local:
	CGO_ENABLED=0  go build -a -installsuffix cgo -v -ldflags "-X ${PROJECT_NAME}/pkg/logging.ProjectName=${PROJECT_NAME}" -o ${TARGET}
build:
	docker run --rm -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE} make local
image: build
	docker build --rm -t ${IMAGE_NAME}:${MAJOR_VERSION}-${GIT_VERSION} .
	docker tag  ${IMAGE_NAME}:${MAJOR_VERSION}-${GIT_VERSION} ${IMAGE_NAME}:${MAJOR_VERSION}
push: image
	docker push ${IMAGE_NAME}:${MAJOR_VERSION}-${GIT_VERSION}
	docker push ${IMAGE_NAME}:${MAJOR_VERSION}

shell:
	docker build --rm -t ${BUILD_IMAGE} contrib/builder/binary
	docker run -ti --rm -v $(shell pwd):/go/src/${PROJECT_NAME} -w /go/src/${PROJECT_NAME} ${BUILD_IMAGE} /bin/bash



.PHONY: image build build-local
