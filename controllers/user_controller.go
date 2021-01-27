package controllers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"code.xxxxx.cn/platform/auth/consts"
	"code.xxxxx.cn/platform/auth/models/db"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils/pagination"
	"golang.org/x/crypto/bcrypt"
)

var (
	specialClients = []int{}
)

func init() {
	ids := strings.Split(beego.AppConfig.String("special_clients"), ",")
	for _, id := range ids {
		if id == "" {
			continue
		}

		id, err := strconv.Atoi(id)
		if err != nil {
			panic("special Clients should have be valid client id")
		}
		specialClients = append(specialClients, id)
	}
}

type UserController struct {
	BaseController
}

func (c *UserController) GetUser() {
	if !c.checkScope("user:read") {
		return
	}
	authInfo := c.getAuthInfo()
	userId := authInfo.UserId

	allowed := false
	if authInfo.LoginType == consts.AUTH_BY_SESSION || authInfo.LoginType == consts.AUTH_BY_TOKEN {
		allowed = true
	}
	if allowed {
		o := orm.NewOrm()
		var user db.User
		if err := o.QueryTable("user").Filter("id", userId).One(&user); err != nil {
			c.Failed(err)
		} else {
			user.Password = ""
			c.Ok(user)
		}
	} else {
		c.InvalidParams(fmt.Errorf("can't read other user data"))
	}
}

func (c *UserController) GetMemberUsers() {
	if !c.checkScope("user:read") {
		return
	}
	//authInfo := c.getAuthInfo()
	//if authInfo.LoginType != consts.AUTH_BY_SESSION {
	//	c.Failed(errors.New("access token or client_id client_secret can't request this api"))
	//	return
	//}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	if !c.checkIsRoleAdminOrSuper(clientId) {
		return
	}
	// FIXME
	if users, err := db.GetAllUsers(false); err != nil {
		c.Failed(err)
	} else {
		c.Ok(users)
	}

}

func (c *UserController) GetUserEncryptedPwds() {
	if !c.checkScope("user:read") {
		return
	}
	var userIds []string
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &userIds); err != nil {
		c.Failed(err)
		return
	}
	users, err := db.GetUserEncryptedPwds(userIds)
	if err != nil {
		c.Failed(err)
		return
	}
	c.Ok(users)
}

func (c *UserController) GetAllUsers() {
	if !c.checkScope("user:read") {
		return
	}
	authInfo := c.getAuthInfo()
	onlyReturnUsers := false
	returnPwd := true
	if authInfo.LoginType == consts.AUTH_BY_SESSION {
		onlyReturnUsers = true
		returnPwd = false
	}

	pageSize, err := c.GetInt("page_size")
	if err != nil {
		pageSize = 10
	}
	filters := make(map[string]string)
	filters["id"] = c.GetString("id")
	filters["fullname"] = c.GetString("fullname")
	filters["status"] = c.GetString("status")
	paginator := pagination.SetPaginator(c.Ctx, pageSize, db.CountUsers(filters))
	users, err := db.GetUsersByOffsetAndLimit(paginator.Offset(), pageSize, returnPwd, filters)
	if err != nil {
		c.Failed(err)
		return
	}
	page := c.PageData(paginator, users)
	if onlyReturnUsers {
		c.Ok(page)
		return
	}

	clientId := c.checkClientId()
	if clientId == 0 {
		c.Ok(page)
		return
	}
	if !c.checkIsRoleAdminOrSuper(clientId) {
		c.Ok(page)
		return
	}
	// get all roles for each user
	for _, user := range users {
		roles, err := db.GetUserRoleByClient(clientId, user.Id, db.RoleNormal, false, false, false, false, false)
		if err != nil {
			logs.Error("Get user role by client failed: ", err)
		}
		user.Roles = roles
	}
	c.Ok(page)
}

func (c *UserController) CreateUser() {
	if !c.checkScope("user:write") {
		return
	}
	authInfo := c.getAuthInfo()
	if authInfo.LoginType == consts.AUTH_BY_SECRET {
		special := false
		for _, id := range specialClients {
			if authInfo.ClientId == int64(id) {
				special = true
				break
			}
		}
		if !special {
			c.Failed(errors.New("auth by client_id and client_secret can't create new client"))
			return
		}
	}
	user := db.User{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &user); err != nil {
		c.Failed(err)
		return
	}
	user.Type = db.UserTypeSys
	user.Status = db.UserStatusActive
	user.Dn = "system user"
	if user.Id == "" {
		c.Failed(errors.New("id cannot be empty"))
		return
	}
	if user.Fullname == "" {
		c.Failed(errors.New("fullname cannot be empty"))
		return
	}
	if user.Password == "" {
		c.Failed(errors.New("password cannot be empty"))
		return
	}
	if user.Organization == "" {
		c.Failed(errors.New("organization cannot be empty"))
		return
	}
	if match, _ := regexp.MatchString(`^([0-9A-Za-z\-_\.]+)@([0-9a-z]+\.[a-z]{2,3}(\.[a-z]{2})?)$`, user.Email); !match {
		c.Failed(errors.New("email format invalid"))
		return
	}

	if res, err := db.CreateUser(user); err != nil {
		c.Failed(err)
	} else {
		c.Ok(res)
	}
}

func (c *UserController) UpdateUserStatus() {
	if !c.checkScope("user:write") {
		return
	}
	authInfo := c.getAuthInfo()
	if authInfo.LoginType == consts.AUTH_BY_SECRET {
		special := false
		for _, id := range specialClients {
			if authInfo.ClientId == int64(id) {
				special = true
				break
			}
		}
		if !special {
			c.Failed(errors.New("auth by client_id and client_secret can't create new client"))
			return
		}
	}
	user := db.User{}
	userId := c.Ctx.Input.Param(":user_id")
	userStatus := c.Ctx.Input.Param(":status")
	if userId == "" {
		c.Failed(errors.New("user_id can't be empty"))
	}
	if userStatus == "" {
		c.Failed(errors.New("user status can't be empty"))
	}
	if userStatus != db.UserStatusActive &&
		userStatus != db.UserStatusFrozen &&
		userStatus != db.UserStatusDelete {
		c.Failed(errors.New("user status invalid. choose one of active,frozen,delete"))
	}
	user.Id = userId
	user.Status = userStatus
	if res, err := db.Update(user); err != nil {
		c.Failed(err)
	} else {
		c.Ok(res)
	}
}

func (c *UserController) UpdateUser() {
	if !c.checkScope("user:write") {
		return
	}
	authInfo := c.getAuthInfo()

	if authInfo.LoginType == consts.AUTH_BY_SECRET {
		special := false
		for _, id := range specialClients {
			if authInfo.ClientId == int64(id) {
				special = true
				break
			}
		}
		if !special {
			c.Failed(errors.New("auth by client_id and client_secret can't create new client"))
			return
		}
	}

	user := db.User{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &user); err != nil {
		c.Failed(err)
	}

	userId := c.Ctx.Input.Param(":user_id")
	user.Id = userId
	if len(user.Password) > 0 {
		pwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.Failed(err)
		} else {
			user.Password = base64.StdEncoding.EncodeToString(pwd)
		}
	}
	if res, err := db.Update(user); err != nil {
		c.Failed(err)
	} else {
		c.Ok(res)
	}
}
