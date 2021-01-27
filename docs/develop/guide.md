# 开发文档

## 模块介绍

项目主体是一个go项目，其中包括前端部分，前端部分在web目录下

## 分支管理

**使用传统gitflow模式管理**

master 为主分支，
各个开发分支开发完毕后先合并分支commit，然后rebase合并到dev分支，dev向master提出MR请求来合并，合并完后打tag

## 项目构建

### 构建镜像

命令 `$ make image`, 使用tag指定镜像版本。

## 项目测试

命令 `$ make run`, 本地跑代码

## 构建并推送镜像

命令 `$ make image-push`
