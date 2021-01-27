# build runtime image
FROM ubuntu:18.04
LABEL maintainer="rpzhang@xxxxx.cn"

WORKDIR /go/src/code.xxxxx.cn/platform/auth

EXPOSE 9096

# 等前后端分离以后在处理这些东西
COPY --from=auth.xxxxx:febuild /src/code.xxxxx.cn/platform/auth/build/ static/
COPY --from=auth.xxxxx:bebuild /go/src/code.xxxxx.cn/platform/auth/auth .
COPY --from=auth.xxxxx:bebuild /go/src/code.xxxxx.cn/platform/auth/views views/
COPY deploy/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip

CMD ["./auth"]
