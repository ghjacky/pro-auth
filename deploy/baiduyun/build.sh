#!/usr/bin/env bash

docker build -t harbor.xxxxx.cn/platform/auth:1.0.3 -f Dockerfile ../../.
docker push harbor.xxxxx.cn/platform/auth:1.0.3
