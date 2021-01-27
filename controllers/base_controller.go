package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"code.xxxxx.cn/platform/auth/consts"
	"code.xxxxx.cn/platform/auth/errors"
	"code.xxxxx.cn/platform/auth/filter"
	"code.xxxxx.cn/platform/auth/models/db"
	"code.xxxxx.cn/platform/auth/service"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils/pagination"
	"golang.org/x/crypto/bcrypt"
)

type TestUser struct {
	Id       string
	Password string
	FullName string
}

var (
	specialUsers = make(map[string]*TestUser)
)

type Response struct {
	ResCode int         `json:"res_code"` // 0-成功  1-参数不合法  2-其他错误
	ResMsg  string      `json:"res_msg"`
	Data    interface{} `json:"data"`
}

type PageData struct {
	Paginator *Paginator  `json:"paginator"`
	Data      interface{} `json:"data"`
}

type Paginator struct {
	PageSize    int   `json:"page_size"`
	TotalSize   int64 `json:"total_size"`
	TotalPages  int   `json:"total_pages"`
	CurrentPage int   `json:"current_page"`
}

func init() {
	users := strings.Split(beego.AppConfig.String("special_users"), ",")
	for _, user := range users {
		tmp := strings.Split(user, ":")
		if len(tmp) < 3 {
			logs.Warning("Special users format invalid, skip all special users account")
			return
		}
		specialUsers[tmp[0]] = &TestUser{
			Id:       tmp[0],
			Password: tmp[1],
			FullName: tmp[2],
		}
	}
}

type BaseController struct {
	beego.Controller
}

func (c *BaseController) InvalidParams(err error) {
	if c.Ctx.Input.Param(":v") == "2" {
		errors.ErrorResponse(err, c.Ctx, http.StatusBadRequest)
	} else {
		msg := "invalid parameters"
		if err != nil {
			msg = err.Error()
		}
		c.Data["json"] = &Response{consts.InvalidParameters, msg, nil}
		c.ServeJSON()
	}

}

func (c *BaseController) PageData(paginator *pagination.Paginator, data interface{}) *PageData {
	return &PageData{
		Paginator: &Paginator{
			PageSize:    paginator.PerPageNums,
			TotalSize:   paginator.Nums(),
			TotalPages:  paginator.PageNums(),
			CurrentPage: paginator.Page(),
		},
		Data: data,
	}
}

func (c *BaseController) Ok(data interface{}) {
	if c.Ctx.Input.Param(":v") == "2" {
		c.Data["json"] = &data
		c.ServeJSON()
	} else {
		c.Data["json"] = &Response{consts.Ok, "ok", &data}
		c.ServeJSON()
	}

}

func (c *BaseController) Failed(err error, statusCode ...int) {
	if c.Ctx.Input.Param(":v") == "2" {
		errors.ErrorResponse(err, c.Ctx, statusCode...)
	} else {
		c.Data["json"] = &Response{consts.InternalError, err.Error(), nil}
		c.ServeJSON()
	}

}

func (c *BaseController) Json(data interface{}) {
	c.Data["json"] = &data
	c.ServeJSON()
}

/**
get AuthInfo from request
*/

func (c *BaseController) getAuthInfo() *filter.AuthInfo {
	if authInfo, ok := c.Ctx.Input.GetData(consts.AUTH_KEY).(filter.AuthInfo); ok {
		return &authInfo
	} else {
		return &filter.AuthInfo{LoginType: consts.NO_AUTH, UserId: "anonymous", ClientId: 0}
	}
}

func (c *BaseController) checkRootRoleSuper(clientId int64) (pass bool) {
	authInfo := c.getAuthInfo()
	if authInfo.LoginType == consts.AUTH_BY_SECRET {
		if authInfo.ClientId == clientId {
			pass = true
		} else {
			c.Failed(errors.ErrInvalidRequest)
			pass = false
		}
	} else {
		o := orm.NewOrm()
		var roleUser []*db.RoleUser
		if num, err := o.Raw("select role_user.* from role_user inner join role on role_user.role_id = role.id where user_id = ? and client_id = ? and parent_id = -1 and role_type = ?", authInfo.UserId, clientId, db.RoleSuper).QueryRows(&roleUser); err != nil {
			logs.Error(err)
			c.Failed(err)
			pass = false
		} else {
			if num > 0 {
				pass = true
			} else {
				c.Failed(errors.ErrUnauthorizedRequest)
				pass = false
			}
		}
	}
	return

}

func (c *BaseController) checkRoleRoot(roleId int64) (isRootRole bool) {
	role, err := db.GetRole(roleId)
	if err != nil {
		return false
	}
	return role.ParentId == -1
}

func (c *BaseController) checkRoleType(clientId, roleId int64, roleType int, judgeParent bool) (pass, isRootRoleSuper bool) {
	authInfo := c.getAuthInfo()
	if authInfo.LoginType == consts.AUTH_BY_SECRET {
		if authInfo.ClientId == clientId {
			pass = true
			isRootRoleSuper = true
		} else {
			c.Failed(errors.ErrUnauthorizedRequest)
			pass = false
			isRootRoleSuper = false
		}
	} else {
		if roles, err := db.GetUserRoleByClient(clientId, authInfo.UserId, roleType, true, false, false, false, false); err != nil {
			logs.Error(err)
			c.Failed(err)
			pass = false
			isRootRoleSuper = false
		} else {
			idMap := make(map[int64]*db.Role)
			for _, v := range roles {
				idMap[v.Id] = v
				if v.ParentId == -1 && v.RoleType == db.RoleSuperStr {
					isRootRoleSuper = true
				}
			}
			if role, ok := idMap[roleId]; ok {
				if judgeParent && role.ParentId != -1 {
					_, isParentExisted := idMap[role.ParentId]
					pass = isParentExisted
					if !isParentExisted {
						c.Failed(errors.ErrUnauthorizedRequest)
					}
				} else {
					pass = true
				}
			} else {
				c.Failed(errors.ErrUnauthorizedRequest)
				pass = false
			}
		}
	}
	return
}

func (c *BaseController) checkIsRoleAdminOrSuper(clientId int64) (pass bool) {
	authInfo := c.getAuthInfo()
	if authInfo.LoginType == consts.AUTH_BY_SECRET {
		if authInfo.ClientId == clientId {
			pass = true
		} else {
			c.Failed(errors.ErrUnauthorizedRequest)
			pass = false
		}
	} else {
		if roles, err := db.GetUserRoleByClient(clientId, authInfo.UserId, db.RoleAdmin, false, false, false, false, false); err != nil {
			logs.Error(err)
			c.Failed(err)
			pass = false
		} else if len(roles) > 0 {
			pass = true
		} else {
			c.Failed(errors.ErrUnauthorizedRequest)
			pass = false
		}
	}
	return
}

/**
check client_id
*/
func (c *BaseController) checkClientId() int64 {
	authInfo := c.getAuthInfo()
	if c.GetString("client_id") == "" && authInfo.ClientId != 0 {
		return authInfo.ClientId
	} else if clientId, err := strconv.ParseInt(c.GetString("client_id"), 10, 64); err != nil || clientId <= 0 {
		c.Failed(errors.ErrInvalidRequest)
		return 0
	} else {
		authInfo := c.getAuthInfo()
		if authInfo.LoginType == consts.AUTH_BY_SESSION {
			return clientId
		} else if authInfo.ClientId == clientId {
			return clientId
		} else {
			c.Failed(errors.ErrInvalidRequest)
			return 0
		}
	}
}

/**
check scope
*/
func (c *BaseController) checkScope(scope string) bool {
	authInfo := c.getAuthInfo()
	if strings.Contains(authInfo.Scope, "all:all") {
		return true
	} else if strings.Contains(authInfo.Scope, scope) {
		return true
	} else {
		c.Failed(errors.ErrUnauthorizedRequest)
		return false
	}
}

/**
check if AuthInfo can read or write the data of user_id

*/
func (c *BaseController) checkUserId(clientId int64) (string, bool) {
	authInfo := c.getAuthInfo()
	userId := c.GetString("user_id")
	if authInfo.LoginType == consts.AUTH_BY_SESSION || authInfo.LoginType == consts.AUTH_BY_TOKEN {
		if authInfo.UserId == userId || userId == "" {
			return authInfo.UserId, true
		} else {
			c.Failed(errors.ErrUnauthorizedRequest)
			return "", false
		}
	} else if authInfo.LoginType == consts.AUTH_BY_SECRET {
		if authInfo.ClientId == clientId {
			if userId == "" {
				c.Failed(errors.ErrInvalidRequest)
				return "", false
			} else {
				return userId, true
			}
		} else {
			c.Failed(errors.ErrUnauthorizedRequest)
			return "", false
		}
	} else {
		c.Failed(errors.ErrUnauthorizedRequest)
		return "", false
	}
}

func (c *BaseController) Login() {
	state := c.GetString("state")
	if c.Ctx.Input.IsGet() {
		if _, ok := c.Ctx.Input.CruSession.Get(consts.SESSION_USER_KEY).(db.User); ok {
			if state == "" {
				state = "/frontend"
			}
			c.Redirect(state, http.StatusFound)
			return
		}

		var loginURL string
		if state != "" {
			loginURL = fmt.Sprintf("/auth/login?state=%s", state)
		} else {
			loginURL = "/auth/login"
		}

		c.Data["loginUrl"] = loginURL
		c.TplName = "login.html"
		return
	}

	if banTime, ok := c.Ctx.Input.CruSession.Get(consts.SESSION_BAN_TIME).(int64); ok && time.Now().Unix() < banTime {
		c.renderLoginPage("", "失败次数过多，请稍后再试", state)
		return
	}

	if !c.Ctx.Input.IsPost() {
		return
	}

	isLogin := false
	var loginErr error
	c.Ctx.Request.Referer()

	retryCount := 0
	if count, ok := c.Ctx.Input.CruSession.Get(consts.SESSION_RETRY_KEY).(int); ok {
		retryCount = count
	}

	authinfo := c.Ctx.Input.Query("authinfo")
	data, err := base64.StdEncoding.DecodeString(authinfo)
	if err != nil {
		logs.Error("Decoding authinfo Failed:%v", err)
		c.Ctx.WriteString("Decode authinfo Failed")
		return
	}
	type Uinfo struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Captcha  string `json:"captcha"`
	}
	u := Uinfo{}
	if err := json.Unmarshal(data, &u); err != nil {
		logs.Error("Unmarshal authinfo Failed:%v", err)
		c.Ctx.WriteString("Decode authinfo Failed")
		return
	}

	captcha := u.Captcha
	if ok := service.Captcha.Verify(consts.SESSION_CAPTCHA_KEY, captcha, c.Ctx); !ok {
		loginErr = fmt.Errorf("验证码错误")
	} else {
		if loginErr = DoLogin(c.Ctx, u.Username, u.Password); loginErr == nil {
			isLogin = true
			retryCount = -1
		}
	}

	if retryCount >= consts.MAX_RETRY_COUNT {
		c.Ctx.Input.CruSession.Set(consts.SESSION_BAN_TIME, time.Now().Add(consts.MIN_BAN_DURATION*time.Minute).Unix())
		retryCount = -1
		loginErr = fmt.Errorf("失败次数过多，请稍后再试")
	} else if loginErr != nil {
		loginErr = fmt.Errorf("%v, 剩余重试次数: %v", loginErr, consts.MAX_RETRY_COUNT-retryCount-1)
	}

	c.Ctx.Input.CruSession.Set(consts.SESSION_RETRY_KEY, retryCount+1)

	if isLogin {
		if state == "" {
			state = "/"
		}
		c.Redirect(state, http.StatusFound)
		return
	}

	c.renderLoginPage(u.Username, loginErr.Error(), c.Ctx.Input.URI())
}

func (c *BaseController) renderLoginPage(username, errMsg, url string) {
	c.Data["username"] = username
	c.Data["password"] = ""
	c.Data["errMsg"] = errMsg
	c.Data["loginUrl"] = url
	c.TplName = "login.html"
	return
}

func (c *BaseController) Logout() {
	sessId := c.Ctx.Input.CruSession.SessionID()
	if err := db.TokenClearBySessionId(sessId); err != nil {
		logs.Error(err)
	}
	if err := c.Ctx.Input.CruSession.Flush(); err != nil {
		logs.Error(err)
	}
	c.Redirect("/auth/login", http.StatusFound)
}

func (c *BaseController) Frontend() {
	if _, ok := c.Ctx.Input.CruSession.Get(consts.SESSION_USER_KEY).(db.User); !ok {
		c.Redirect(fmt.Sprintf("/auth/login?state=%s", c.Ctx.Input.URI()), http.StatusFound)
	} else {
		c.ViewPath = "static"
		c.TplName = "index.html"
	}
}

func (c *BaseController) Captcha() {
	err := service.Captcha.NewImageToSession(consts.SESSION_CAPTCHA_KEY, c.Ctx)
	if err != nil {
		logs.Error(err)
		c.Data["json"] = err
		c.ServeJSON()
	}
}

func (c *BaseController) Echo() {
	c.Data["json"] = c.Ctx.Input.IP()
	c.ServeJSON()
}

func DoLogin(ctx *context.Context, username, password string) error {
	var userId string
	var user db.User

	if ok, _ := regexp.MatchString(fmt.Sprintf(`^[a-zA-Z0-9\_\-]+%s$`, beego.AppConfig.String("ldap_mailPostfix")), username); ok {
		userId = strings.Split(username, "@")[0]
	} else if ok, _ = regexp.MatchString(`^[a-zA-Z0-9\_\-]+$`, username); ok {
		userId = username
	} else {
		logs.Error("username format invalid: not match a-zA-Z0-9_-")
		return fmt.Errorf("用户名或密码错误")

	}

	if isValid, err := IsUserValid(userId, password); err != nil {
		return err
	} else if !isValid {
		return fmt.Errorf("用户名或密码错误")
	}

	if u, err := QueryAndCreateUser(userId, password); err != nil {
		logs.Error("query and create user function error: %v", err)
		return fmt.Errorf("用户名或密码错误")
	} else {
		user = *u
	}

	if err := ctx.Input.CruSession.Set(consts.SESSION_USER_KEY, user); err != nil {
		return err
	}
	return nil
}

func IsUserValid(userId, password string) (bool, error) {
	var isValid bool
	var mail = userId + beego.AppConfig.String("ldap_mailPostfix")
	if u, exist := specialUsers[userId]; exist {
		pwd, err := base64.StdEncoding.DecodeString(u.Password)
		isValid = (err == nil) && bcrypt.CompareHashAndPassword(pwd, []byte(password)) == nil
	} else {
		if service.Ldap.Enabled() {
			if ok, err := service.Ldap.Auth(mail, password); ok && err == nil {
				isValid = true
			}
		}

		if !isValid {
			if persisUser, err := QueryUser(userId); err == nil {
				if persisUser.Type == db.UserTypeLdap {
					isValid = false
					return isValid, nil
				}
				pwd, err := base64.StdEncoding.DecodeString(persisUser.Password)
				if err != nil {
					isValid = false
					return isValid, nil
				}
				err = bcrypt.CompareHashAndPassword(pwd, []byte(password))
				if err != nil {
					isValid = false
					return isValid, nil
				}
				if persisUser.Status != db.UserStatusActive {
					isValid = false
					return isValid, errors.ErrUserNotActive
				}
				isValid = true
				return isValid, nil
			}
		}
	}

	if isValid {
		if _, err := QueryAndCreateUser(userId, password); err != nil {
			return false, err
		}
	}
	return isValid, nil
}

func QueryUser(userId string) (*db.User, error) {
	var persisUser db.User
	o := orm.NewOrm()
	if err := o.QueryTable("user").Filter("id", userId).One(&persisUser); err != nil {
		return nil, err
	}
	return &persisUser, nil
}

func QueryAndCreateUser(userId, password string) (*db.User, error) {
	if len(userId) == 0 {
		return nil, errors.ErrUserNotFound
	}
	var mailPrefix = ""
	if service.Ldap.Enabled() {
		mailPrefix = beego.AppConfig.String("ldap_mailPostfix")
	} else {
		mailPrefix = "@" + beego.AppConfig.String("org_domain")
	}
	var mail = userId + mailPrefix
	var user db.User

	user.Id = userId
	pwd, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user.Password = base64.StdEncoding.EncodeToString(pwd)
	user.Email = mail
	user.Type = db.UserTypeSys
	user.Dn = "system user"
	user.Status = db.UserStatusActive

	var persisUser db.User
	if u, ok := specialUsers[userId]; ok {
		user.Fullname = u.FullName
		user.Type = db.UserTypeSpecial
		user.Dn = "special user"
	} else if service.Ldap.Enabled() {
		if dn, cnName, err := service.Ldap.GetUserInfo(mail); err == nil {
			user.Fullname = cnName
			user.Type = db.UserTypeLdap
			user.Dn = dn
		}
	}

	o := orm.NewOrm()
	if err := o.QueryTable("user").Filter("id", userId).One(&persisUser); err != nil {
		if err == orm.ErrNoRows {
			if _, err := o.Insert(&user); err != nil {
				return nil, err
			}
			return &user, nil
		}
		return nil, errors.ErrUserNotFound
	}

	// update ldap dn
	params := orm.Params{}

	if persisUser.Dn != user.Dn {
		persisUser.Dn = user.Dn
		params["dn"] = user.Dn
	}

	if pwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); len(password) > 0 && err == nil {
		base64Pwd := base64.StdEncoding.EncodeToString(pwd)
		if persisUser.Password != base64Pwd {
			persisUser.Password = base64Pwd
			params["password"] = base64Pwd
		}
	}
	params["type"] = user.Type
	if len(params) > 0 {
		if _, err := o.QueryTable("user").Filter("id", userId).Update(params); err != nil {
			logs.Error(err)
		}
	}

	return &persisUser, nil
}
