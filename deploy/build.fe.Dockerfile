 FROM node:11-alpine

LABEL maintainer="rpzhang@xxxxx.cn"

# Workdir
WORKDIR $GOPATH/src/code.xxxxx.cn/platform/auth

COPY web/package-bkp.json ./package.json
COPY web/package-lock-bkp.json ./package-lock.json

RUN npm install --registry=https://registry.npm.taobao.org

COPY web/ .

# npm build
RUN npm install --registry=https://registry.npm.taobao.org && npm run build
