package main

import (
	"code.xxxxx.cn/platform/auth/models/db"
	_ "code.xxxxx.cn/platform/auth/routers"
	"code.xxxxx.cn/platform/auth/service"
	"github.com/astaxie/beego/orm"

	"encoding/gob"
	"net/url"

	_ "code.xxxxx.cn/platform/auth/session/redis_sentinel"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
)

func init() {
	if beego.BConfig.RunMode == "dev" {
		orm.Debug = true
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	beego.BConfig.WebConfig.Session.SessionGCMaxLifetime = 3600 * 8
	beego.BConfig.WebConfig.Session.SessionCookieLifeTime = 3600 * 8
}

func main() {
	gob.Register(db.User{})
	gob.Register(service.CaptchaValue{})

	beego.AddViewPath("static")
	logs.SetLogFuncCallDepth(3)
	beego.SetLogger("file", `{"filename":"logs/auth.log","level":7,"maxlines":0,"maxsize":0,"daily":true,"maxdays":30}`)

	beego.InsertFilter("/*", beego.BeforeRouter, func(context *context.Context) {
		if context.Input.URL() == "/echo" {
			return
		}
		uri, _ := url.PathUnescape(context.Input.URI())
		logs.Info("method: %s, uri: %s", context.Input.Method(), uri)
		if context.Input.URL() == "/oauth2/token" || context.Input.URL() == "/auth/login" || context.Input.URL() == "/oauth2/authorize" {
			return
		}
		contentType := context.Input.Header("Content-Type")
		if contentType == "application/x-www-form-urlencoded" {
			logs.Info("body: %v", context.Request.Form)
		} else if contentType == "application/json" {
			logs.Info("body: %v", context.Input.RequestBody)
		}
		logs.Info("beegosessionID: %s", context.GetCookie("beegosessionID"))
	})
	beego.Run()
}
