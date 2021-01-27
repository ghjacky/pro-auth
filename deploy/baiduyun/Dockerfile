FROM golang:1.9.4
LABEL maintainer="xincaigu@xxxxx.cn"

WORKDIR $GOPATH/src/code.xxxxx.cn/platform/auth
COPY . .

RUN set -ex \
    && apt-get update \
    && apt-get install -y curl \
    && curl -sL https://deb.nodesource.com/setup_11.x | bash - \
    && apt-get install -y nodejs \
    && go build \
    && rm -rf logs && mkdir logs \
    && cd web && npm install --registry=https://registry.npm.taobao.org && npm run build \
    && cd .. \
    && rm -rf static \
    && cp -r web/build/ static/

#COPY entrypoint.sh /entrypoint.sh

EXPOSE 9096

#ENTRYPOINT ["/entrypoint.sh"]

# set default arg for entrypoint
CMD ["./auth"]
