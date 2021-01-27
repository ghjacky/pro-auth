FROM golang:1.12 as build
LABEL maintainer="yjyu@xxxxx.cn"
WORKDIR $GOPATH/src/code.xxxxx.cn/platform/auth
COPY . .
RUN go build -o user-sync -v ./sync/

FROM ubuntu:18.04
LABEL maintainer="yjyu@xxxxx.cn"
WORKDIR /usr/local/bin
COPY --from=build /go/src/code.xxxxx.cn/platform/auth/user-sync .
COPY --from=build /go/src/code.xxxxx.cn/platform/auth/sync/conf conf/
COPY deploy/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
CMD ["./user-sync"]
