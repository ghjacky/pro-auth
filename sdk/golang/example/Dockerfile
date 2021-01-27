FROM golang:1.12 as build

LABEL maintainer="rpzhang@xxxxx.cn"

# Env prepare
RUN go get -u github.com/golang/dep/cmd/dep

# Workdir
WORKDIR $GOPATH/src/code.xxxxx.cn/platform/auth

COPY . .

# Go build
RUN go build -v sdk/golang/example/example.go

FROM ubuntu:18.04

LABEL maintainer="rpzhang@xxxxx.cn"

WORKDIR /usr/local/bin

# 等前后端分离以后在处理这些东西
COPY --from=build /go/src/code.xxxxx.cn/platform/auth/example .

CMD ["./example"]
