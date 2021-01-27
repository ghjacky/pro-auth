package controllers

import (
	"fmt"

	"code.xxxxx.cn/platform/auth/consts"
	"code.xxxxx.cn/platform/auth/errors"
	"code.xxxxx.cn/platform/auth/models/db"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type AuthController struct {
	BaseController
}

func (c *AuthController) CheckResource() {
	authInfo := c.getAuthInfo()
	if authInfo.LoginType != consts.AUTH_BY_SECRET {
		c.Failed(errors.ErrWrongAuth)
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	resourceName := c.GetString("resource_name")
	userId := c.GetString("user_id")

	sql := "select role_user.user_id, role_resource.resource_id, resource.name  from role_user join role_resource on role_user.role_id = role_resource.role_id join resource on role_resource.resource_id = resource.id where role_user.user_id = ? and resource.name = ?"

	o := orm.NewOrm()
	var lists []orm.ParamsList
	if num, err := o.Raw(sql, userId, resourceName).ValuesList(&lists); err != nil {
		logs.Error(err)
		c.Ok(false)
	} else if num == 0 {
		c.Ok(false)
	} else {
		c.Ok(true)
	}
}

func (c *AuthController) AuthUser() {
	authInfo := c.getAuthInfo()
	if authInfo.LoginType != consts.AUTH_BY_SECRET {
		c.Failed(errors.ErrWrongAuth)
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}

	username := c.Ctx.Input.Query("username")
	password := c.Ctx.Input.Query("password")

	if err := DoLogin(c.Ctx, username, password); err != nil {
		c.Failed(err)
	} else {
		c.Ok("auth succeed")
	}
}

func (c *AuthController) CheckThirdParty() {
	access := c.GetString("access_token")
	userID := c.GetString("user_id")
	clientID := c.GetString("client_id")

	store := db.TokenStore{}
	token, err := store.GetByAccess(access)
	if err != nil {
		errors.ErrorResponse(err, c.Ctx)
		return
	}

	if token.IsAccessExpired() {
		errors.ErrorResponse(fmt.Errorf("access token expired"), c.Ctx)
		return
	}

	c.Ok(token.GetClientId() == clientID && token.GetUserId() == userID)
}

func (c *AuthController) GenerateUserSecret() {
	info := c.getAuthInfo()
	if info.UserId == "" {
		return
	}

	secret := RandStringBytesMaskImprSrcUnsafe(20)

	user := db.User{
		Id:     info.UserId,
		Secret: secret,
	}

	_, err := db.Update(user)
	if err != nil {
		c.Failed(err)
	}
	c.Ok("success")
}
