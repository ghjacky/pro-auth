package controllers

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"code.xxxxx.cn/platform/auth/consts"
	"code.xxxxx.cn/platform/auth/errors"
	"code.xxxxx.cn/platform/auth/filter"
	"code.xxxxx.cn/platform/auth/models/db"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/dgrijalva/jwt-go"
	"github.com/sagacioushugo/oauth2"
	"github.com/sagacioushugo/oauth2/server"
	"github.com/sagacioushugo/oauth2/store"
)

var Srv *server.Server

func init() {
	config := oauth2.NewDefaultConfig()
	config.ManagerConfig.TokenStoreName = "mysql"
	config.AllowGrantType[oauth2.Implicit] = oauth2.GrantTypeConfig{
		AccessTokenExpire:  6 * 3600,
		RefreshTokenExpire: 24 * 3600,
		IsGenerateRefresh:  true,
		IsResetRefreshTime: false,
	}
	config.AllowGrantType[oauth2.PasswordCredentials] = oauth2.GrantTypeConfig{
		AccessTokenExpire:  6 * 3600,
		RefreshTokenExpire: 24 * 3600,
		IsGenerateRefresh:  true,
		IsResetRefreshTime: false,
	}
	Srv = server.NewServer(config)
	go Srv.Manager.TokenGC()

	Srv.SetCheckUserPasswordHandler(checkUserPasswordHandler)
	Srv.SetCheckUserGrantAccessHandler(checkUserGrantAccessHandler)
	Srv.SetAuthenticateClientHandler(authenticateClientHandler)
	Srv.SetCustomizedCheckScopeHandler(customizedCheckScopeHandler)
	Srv.SetCustomizedAuthorizeErrHandler(server.ResponseErr)

}

type Oauth2Controller struct {
	BaseController
}

func (c *Oauth2Controller) Authorize() {
	// TODO: replace sagacioushugo's server to github.com/golang/oauth2
	Srv.Authorize(c.Ctx)
}

func (c *Oauth2Controller) Token() {
	// TODO: replace sagacioushugo's server to github.com/golang/oauth2
	authorizationToken := c.Ctx.Input.Header(filter.HEADER_AUTHORIZATION)
	if len(strings.Split(authorizationToken, " ")) > 1 {
		authorizationToken = strings.Split(authorizationToken, " ")[1]
	}

	client, err := filter.ParseJWTToken(authorizationToken)
	if err != nil {
		errors.ErrorResponse(err, c.Ctx)
		return
	}

	basic := fmt.Sprintf("%s:%s", url.QueryEscape(strconv.FormatInt(client.Id, 10)), url.QueryEscape(client.Secret))
	basic = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(basic)))

	c.Ctx.Request.Header.Set("Authorization", basic)

	Srv.Token(c.Ctx)
}

func (c *Oauth2Controller) IsTokenExpired() {
	access := c.GetString("access_token")
	refresh := c.GetString("refresh_token")
	var token store.Token
	var err error
	if access != "" {
		if token, err = Srv.Manager.TokenGetByAccess(access); err != nil {
			errors.ErrorResponse(err, c.Ctx)
			return
		}
	} else if refresh != "" {
		if token, err = Srv.Manager.TokenGetByRefresh(refresh); err != nil {
			errors.ErrorResponse(err, c.Ctx)
			return
		}
	} else {
		errors.ErrorResponse(errors.ErrInvalidRequest, c.Ctx)
		return
	}
	res := make(map[string]bool)
	res["is_access_expired"] = token.IsAccessExpired()
	res["is_refresh_expired"] = token.IsRefreshExpired()
	c.Data["json"] = &res
	c.ServeJSON()

}

func (c *Oauth2Controller) DestroyToken() {
	access := c.GetString("access_token")
	if access == "" {
		errors.ErrorResponse(oauth2.ErrInvalidAccessToken, c.Ctx)
	} else if token, err := Srv.Manager.TokenGetByAccess(access); err != nil {
		errors.ErrorResponse(err, c.Ctx)
	} else if token.IsAccessExpired() {
		errors.ErrorResponse(oauth2.ErrExpiredAccessToken, c.Ctx)

	} else {
		t := token.(*db.Token)
		if err := db.TokenClearBySessionId(t.GrantSessionId); err != nil {
			errors.ErrorResponse(err, c.Ctx)
		} else {
			if err := beego.GlobalSessions.GetProvider().SessionDestroy(t.GrantSessionId); err != nil {
				logs.Error(err)
			}
			c.Ctx.WriteString("access has been destroyed")
		}
	}

}

func checkUserPasswordHandler(jwtToken, _ string, ctx *context.Context) (string, error) {
	if username, err := authenticateUserSecret(jwtToken); err == nil {
		return username, nil
	}

	return "", oauth2.ErrInvalidUsernameOrPassword
}

func checkUserGrantAccessHandler(req *oauth2.Request, ctx *context.Context) (userId string, err error) {
	user, ok := ctx.Input.CruSession.Get(consts.SESSION_USER_KEY).(db.User)
	if ok {
		return user.Id, nil
	}

	c := beego.Controller{}
	c.Init(ctx, "", "", c)
	params := url.Values{
		"state": []string{ctx.Input.URI()},
	}

	c.Data["loginUrl"] = "/auth/login?" + params.Encode()
	c.Data["buttonName"] = "授权并登陆"
	c.TplName = "login.html"
	err = c.Render()
	if err != nil {
		err = oauth2.ErrAccessDenied
	}

	return
}

func authenticateClientHandler(ctx *context.Context, clientIdAndSecret ...string) (redirectUris string, err error) {
	if len(clientIdAndSecret) == 0 || len(clientIdAndSecret) > 2 {
		return "", oauth2.ErrInvalidRequest
	}
	o := orm.NewOrm()
	var client db.Client
	q := o.QueryTable("client").Filter("id", clientIdAndSecret[0])

	if len(clientIdAndSecret) == 2 {
		q.Filter("secret", clientIdAndSecret[1])
	}

	if err := q.One(&client); err != nil {
		if err == orm.ErrNoRows {
			return "", oauth2.ErrInvalidClient
		} else {
			return "", err
		}
	} else {
		data := ctx.Input.Data()
		data["clientName"] = client.Fullname
		return client.RedirectUri, nil
	}
}

func customizedCheckScopeHandler(scope string, grantType *oauth2.GrantType, ctx *context.Context) (allowed bool, err error) {
	var scopeMap = map[string]string{
		"role":          "read|write",
		"role-user":     "read|write",
		"role-resource": "read|write",
		"resource":      "read|write",
		"user":          "read|write",
		"client":        "read|write",
		"groupuser":     "read|write",
		"group":         "read|write",
		"all":           "all",
	}
	arr := strings.Split(scope, " ")
	allowed = true
	for i := range arr {
		cell := strings.Split(arr[i], ":")
		if len(cell) != 2 {
			allowed = false
			break
		}
		v, ok := scopeMap[cell[0]]
		if !ok {
			allowed = false
			break
		}
		if strings.Index(v, cell[1]) == -1 {
			allowed = false
			break
		}
	}
	return allowed, nil
}

func authenticateUserSecret(jwtToken string) (string, error) {
	username := ""
	_, err := jwt.Parse(jwtToken, func(t *jwt.Token) (interface{}, error) {
		var ok bool

		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("Unexpected claims type")
		}

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", claims["username"])
		}

		timeStampString, ok := claims["time"].(string)
		if !ok {
			return nil, fmt.Errorf("Unexpected timeStamp format, it should be string")
		}
		timeStamp, err := strconv.Atoi(timeStampString)
		if err != nil {
			return nil, fmt.Errorf("Unexpected timeStamp length, it should be int64")
		}

		if time.Now().Unix()-int64(timeStamp) > 300 {
			return nil, fmt.Errorf("Unexpected timeStamp in claims: %d", timeStamp)
		}

		username, ok = claims["username"].(string)
		if !ok {
			return nil, fmt.Errorf("Unexpected username: %v", username)
		}

		users, err := db.GetUsersByID([]string{username})
		if err != nil || len(users) == 0 {
			return nil, err
		}

		return []byte(users[0].Secret), nil
	})

	return username, err
}
