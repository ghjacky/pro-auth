FROM golang:1.12 as build
LABEL maintainer="rpzhang@xxxxx.cn"
WORKDIR $GOPATH/src/code.xxxxx.cn/platform/auth
COPY . .
RUN go build -o sync.bin -v ./cmd/syncClients/

FROM alpine:latest
LABEL maintainer="rpzhang@xxxxx.cn"
RUN apk add --no-cache libc6-compat

WORKDIR /opt
COPY --from=build /go/src/code.xxxxx.cn/platform/auth/sync.bin sync
