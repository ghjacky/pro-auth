#!/usr/bin/env bash
docker build -t registry.xxxxx.cn/platform/auth-sync:v1.0.0 .
docker push registry.xxxxx.cn/platform/auth-sync:v1.0.0
