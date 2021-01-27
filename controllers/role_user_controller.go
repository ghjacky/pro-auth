package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"code.xxxxx.cn/platform/auth/errors"
	"code.xxxxx.cn/platform/auth/models/db"
	"github.com/astaxie/beego/orm"
)

type RoleUserController struct {
	BaseController
}

/**
查询角色中的用户
*/
func (c *RoleUserController) GetRoleUser() {
	if !c.checkScope("role-user:read") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	roleId, _ := strconv.ParseInt(c.Ctx.Input.Param(":role_id"), 10, 64)
	if roleId > 0 {
		if pass, _ := c.checkRoleType(clientId, roleId, db.RoleAdmin, false); !pass {
			return
		}
	} else {
		if !c.checkRootRoleSuper(clientId) {
			return
		}
	}
	if roleUsers, err := db.GetRoleUsers(clientId, roleId); err != nil {
		c.Failed(err)
	} else {
		c.Ok(roleUsers)
	}
}

/**
在角色中新增用户（批量）
*/
func (c *RoleUserController) AddRoleUser() {
	if !c.checkScope("role-user:write") {
		return
	}
	clientID := c.checkClientId()
	if clientID == 0 {
		return
	}
	var err error
	var users []db.RoleUserJSON

	roleID, err := strconv.ParseInt(c.Ctx.Input.Param(":role_id"), 10, 64)
	if err != nil {
		c.Failed(err)
		return
	}
	pass, isRootRoleSuper := c.checkRoleType(clientID, roleID, db.RoleAdmin, false)
	if !pass {
		return
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &users); err != nil {
		c.Failed(err)
		return
	}

	for index := range users {
		users[index].RoleId = roleID
	}

	if num, err := c.AddRoleUserBatch(users, isRootRoleSuper); err != nil {
		c.Failed(err)
	} else {
		c.Ok(num)
	}
}

func (c *RoleUserController) AddRoleUserBatchImpl() {
	if !c.checkScope("role-user:write") {
		return
	}

	clientID := c.checkClientId()
	if clientID == 0 {
		return
	}

	var users []db.RoleUserJSON
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &users); err != nil {
		c.Failed(err)
		return
	}

	if num, err := c.AddRoleUserBatch(users, true); err != nil {
		c.Failed(err)
	} else {
		c.Ok(num)
	}
}

func (c *RoleUserController) AddRoleUserBatch(users []db.RoleUserJSON, allowSuperRole bool) (int64, error) {

	var existedUsers []db.User
	var notExistedUsers []db.RoleUserJSON
	var existedMap = map[string]string{}
	var userIds = make([]string, len(users))
	for i, v := range users {
		userIds[i] = v.UserId
		if v.RoleType == db.RoleSuperStr && !allowSuperRole {
			return 0, errors.ErrNoUnauthorizedToRoleTypeSuper
		}
	}
	o := orm.NewOrm()
	if _, err := o.QueryTable("user").Filter("id__in", userIds).All(&existedUsers); err != nil {
		return 0, err
	}

	if len(existedUsers) == 0 {
		return 0, errors.ErrUserNotFound
	}

	if len(users) != len(existedUsers) {
		for _, v := range existedUsers {
			existedMap[v.Id] = v.Id
		}
		errorMap := make(map[string]string)
		for _, v := range notExistedUsers {
			errorMap[v.UserId] = "User Not Exist"
		}
		if len(errorMap) > 0 {
			errorMsg, _ := json.Marshal(errorMap)
			return 0, fmt.Errorf(string(errorMsg))
		}
	}

	return db.AddRoleUserBatch(users)
}

/**
修改用户在某角色中的身份
*/
func (c *RoleUserController) UpdateRoleUser() {
	if !c.checkScope("role-user:write") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	var err error
	var roleId int64
	var user db.RoleUserJSON
	if roleId, err = strconv.ParseInt(c.Ctx.Input.Param(":role_id"), 10, 64); err != nil {
		c.Failed(err)
		return
	}
	authInfo := c.getAuthInfo()
	pass, isRootRoleSuper := c.checkRoleType(clientId, roleId, db.RoleAdmin, false)
	if !pass {
		return
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &user); err != nil {
		c.Failed(err)
		return
	}
	if user.UserId == authInfo.UserId {
		c.InvalidParams(fmt.Errorf("can't modify oneself role type"))
		return
	}
	if _, err := QueryAndCreateUser(user.UserId, ""); err != nil {
		c.Failed(err)
		return
	}
	if res, err := db.UpdateRoleUser(user.UserId, roleId, user.RoleType, isRootRoleSuper); err != nil {
		c.Failed(err)
	} else {
		c.Ok(res)
	}
}

/**
删除角色中的用户（批量）
*/

func (c *RoleUserController) DeleteRoleUser() {
	if !c.checkScope("role-user:write") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}
	var err error
	var roleId int64
	var userIds []string
	if roleId, err = strconv.ParseInt(c.Ctx.Input.Param(":role_id"), 10, 64); err != nil {
		c.Failed(err)
		return
	}
	pass, isRootRoleSuper := c.checkRoleType(clientId, roleId, db.RoleAdmin, false)
	if !pass {
		return
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &userIds); err != nil {
		c.Failed(err)
		return
	}
	authInfo := c.getAuthInfo()
	for _, v := range userIds {
		if v == authInfo.UserId && c.checkRoleRoot(roleId) {
			c.InvalidParams(fmt.Errorf("can't delete oneself from root role"))
			return
		}
	}
	o := orm.NewOrm()
	if num, err := o.QueryTable("role_user").Filter("user_id__in", userIds).Filter("role_id", roleId).Filter("role_type", db.RoleSuper).Count(); err != nil && err != orm.ErrNoRows {
		c.Failed(err)
		return
	} else if num > 0 && !isRootRoleSuper {
		c.Failed(errors.ErrNoUnauthorizedToRoleTypeSuper)
		return
	}
	if num, err := db.DeleteRoleUser(userIds, roleId); err != nil {
		c.Failed(err)
	} else {
		c.Ok(num)
	}
}

/**
删除角色用户关系（批量）
*/
func (c *RoleUserController) DeleteRoleUserBatchImpl() {
	if !c.checkScope("role-user:write") {
		return
	}
	clientId := c.checkClientId()
	if clientId == 0 {
		return
	}

	var users []db.RoleUserJSON
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &users); err != nil {
		c.Failed(err)
		return
	}

	userRole := make(map[string][]int64)
	for _, u := range users {
		if roles, ok := userRole[u.UserId]; ok {
			userRole[u.UserId] = append(roles, u.RoleId)
		} else {
			userRole[u.UserId] = []int64{u.RoleId}
		}
	}

	for userID, roleIds := range userRole {
		if num, err := db.DeleteUserRole(userID, roleIds); err != nil {
			c.Failed(err)
		} else {
			c.Ok(num)
		}
	}
}
