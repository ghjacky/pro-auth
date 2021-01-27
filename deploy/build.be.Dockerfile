FROM golang:1.9.4 

LABEL maintainer="rpzhang@xxxxx.cn"

# Env prepare
#RUN go get -u github.com/golang/dep/cmd/dep

# Workdir
WORKDIR $GOPATH/src/code.xxxxx.cn/platform/auth
COPY . .

# Go build
#RUN dep ensure && go build -v
RUN go build -v
