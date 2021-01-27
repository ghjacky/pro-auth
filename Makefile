# Copyright 2019 xxxxx Inc.
SHELL := /bin/bash

REGISTRY=harbor.xxxxx.cn
REGISTRY_IDC=registry.xxxxx.cn
REPO=${REGISTRY}/platform/authority-platform
VERSION=$(shell git describe --tags --abbrev=7 HEAD --always)
IMAGE_NAME=${REPO}:${VERSION}
PROJECT_PATH=/go/src/code.xxxxx.cn/platform/auth
BEEGO_RUNMODE=dev
RUN_PORT=9096
LOCAL_MYSQL_PORT=3306
LOCAL_MYSQL_PWD=123456
SYNC_VERSION=v1.0.2

.PHONY: image
image:
		docker build --rm . -f deploy/build.be.Dockerfile -t auth.xxxxx:bebuild
		docker build --rm . -f deploy/build.fe.Dockerfile -t auth.xxxxx:febuild
		docker build --rm . -f deploy/deploy.Dockerfile -t ${IMAGE_NAME}

.PHONY: image-push
image-push: image
		docker push ${IMAGE_NAME}
		docker tag ${IMAGE_NAME} ${REGISTRY_IDC}/platform/authority-platform:${VERSION}
		docker push ${REGISTRY_IDC}/platform/authority-platform:${VERSION}

.PHONY: dep-image
dep-image:
		docker build --rm deploy/build/dep -t ${REGISTRY}/platform/dep:latest

.PHONY: dep-image-push
dep-image-push:
		docker push ${REGISTRY}/platform/dep:latest
		docker tag ${REGISTRY}/platform/dep:latest ${REGISTRY_IDC}/platform/dep:latest
		docker push ${REGISTRY_IDC}/platform/dep:latest

.PHONY: vendor
vendor:
		docker run --rm -v $(PWD):${PROJECT_PATH} -w ${PROJECT_PATH} ${REGISTRY}/platform/dep:latest dep ensure -v

.PHONY: gobuild
gobuild: vendor
		docker run --rm -v $(PWD):${PROJECT_PATH} -w ${PROJECT_PATH} golang:1.9.4 go build

.PHONY: node_modules
node_modules:
		docker run --rm -v $(PWD)/web:${PROJECT_PATH} -w ${PROJECT_PATH} node:10.16.3-stretch-slim npm install --registry=https://registry.npm.taobao.org

.PHONY: npm_build
npm_build: node_modules
		docker run --rm -v $(PWD)/web:${PROJECT_PATH} -w ${PROJECT_PATH} node:10.16.3-stretch-slim npm run build

.PHONY: run
run: image
ifeq ($(BEEGO_RUNMODE),local)
		mkdir -p tmpdata logs
		docker run --rm -p ${LOCAL_MYSQL_PORT}:3306 -v $(PWD)/logs:/logs -v $(PWD)/tmpdata:/var/lib/mysql -e MYSQL_ROOT_PASSWORD=${LOCAL_MYSQL_PWD} -d mysql:5.6 & sleep 5
endif
		docker run --rm -v $(PWD)/conf:${PROJECT_PATH}/conf -p ${RUN_PORT}:${RUN_PORT} -e BEEGO_RUNMODE=${BEEGO_RUNMODE} -w ${PROJECT_PATH} ${IMAGE_NAME}

.PHONY: build_example
build_example:
		docker build --rm . -f sdk/golang/example/Dockerfile -t ${REGISTRY}/platform/auth_example:golang
		docker build --rm . -f sdk/python/example/Dockerfile -t ${REGISTRY}/platform/auth_example:python
		docker build --rm . -f sdk/python/example/py3.Dockerfile -t ${REGISTRY}/platform/auth_example:python3

.PHONY: build_sync
build_sync:
		docker build --rm . -f sync/Dockerfile -t ${REGISTRY}/platform/auth-sync:${SYNC_VERSION}
		docker push ${REGISTRY}/platform/auth-sync:${SYNC_VERSION}
		docker tag ${REGISTRY}/platform/auth-sync:${SYNC_VERSION} ${REGISTRY_IDC}/platform/auth-sync:${SYNC_VERSION}
		docker push ${REGISTRY_IDC}/platform/auth-sync:${SYNC_VERSION}

.PHONY: build_clone
build_clone:
		docker build --rm . -f cmd/clone/Dockerfile -t ${REGISTRY}/platform/auth-clone:${VERSION}
		docker push ${REGISTRY}/platform/auth-clone:${VERSION}

.PHONY: build_syncclient
build_syncclient:
		docker build --rm . -f cmd/syncClients/Dockerfile -t ${REGISTRY}/platform/auth-syncclients:${VERSION}
		docker push ${REGISTRY}/platform/auth-syncclients:${VERSION}
